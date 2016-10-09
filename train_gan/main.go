package main

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/gans"
	"github.com/unixpickle/num-analysis/linalg"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
)

const (
	FaceSize  = 28
	StepSize  = 0.001
	BatchSize = 64
)

const (
	ImagesArg = 1
	ModelArg  = 2
	GenArg    = 3
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: train_gan <image_path> <model_out> <gen.png>")
		os.Exit(1)
	}

	samples := readFaceImages()
	model := createModel(samples)

	var iteration int
	sgd.SGDMini(model, samples, StepSize, BatchSize, func(s sgd.SampleSet) bool {
		posCost := model.SampleRealCost(samples)
		genCost := model.SampleGenCost()
		log.Printf("iteration %d: real_cost=%f  gen_cost=%f", iteration,
			posCost, genCost)
		iteration++
		return true
	})

	log.Println("Saving model...")
	data, err := model.Serialize()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(os.Args[ModelArg], data, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Println("Creating generation grid...")
	renderings := gans.GridSample(5, 8, func() *neuralnet.Tensor3 {
		randVec := make([]float64, model.RandomSize)
		for i := range randVec {
			randVec[i] = rand.NormFloat64()
		}
		out := model.Generator.Apply(&autofunc.Variable{Vector: randVec}).Output()
		return &neuralnet.Tensor3{Width: 28, Height: 28, Depth: 3, Data: out}
	})
	outFile, err := os.Create(os.Args[GenArg])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer outFile.Close()
	png.Encode(outFile, renderings)
}

func readFaceImages() sgd.SampleSet {
	path := os.Args[ImagesArg]
	listing, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var res sgd.SliceSampleSet
	for _, item := range listing {
		ext := strings.ToLower(filepath.Ext(item.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		subPath := filepath.Join(path, item.Name())
		f, err := os.Open(subPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %s: %s\n", subPath, err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode %s: %s\n", subPath, err)
			continue
		}
		if img.Bounds().Dx() != FaceSize || img.Bounds().Dy() != FaceSize {
			fmt.Fprintln(os.Stderr, "Bad size for:", subPath)
			continue
		}
		tensor := neuralnet.NewTensor3(FaceSize, FaceSize, 3)
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				r, g, b, _ := img.At(x+img.Bounds().Min.X, y+img.Bounds().Min.Y).RGBA()
				tensor.Set(x, y, 0, float64(r)/0xffff)
				tensor.Set(x, y, 1, float64(g)/0xffff)
				tensor.Set(x, y, 2, float64(b)/0xffff)
			}
		}
		res = append(res, neuralnet.VectorSample{
			Input:  tensor.Data,
			Output: linalg.Vector{},
		})
	}
	return res
}

func createModel(samples sgd.SampleSet) *gans.FM {
	existing, err := ioutil.ReadFile(os.Args[ModelArg])
	if err == nil {
		model, err := gans.DeserializeFM(existing)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to deserialize model:", err)
			os.Exit(1)
		}
		log.Println("Loaded existing model.")
		return model
	}

	log.Println("Created new model.")

	discrim := createDiscriminator(samples)
	return &gans.FM{
		Discriminator: discrim,
		FeatureLayers: len(discrim) - 2,
		Generator:     createGenerator(),
		RandomSize:    14 * 14,
	}
}

func createGenerator() neuralnet.Network {
	var res neuralnet.Network

	for i := 0; i < 2; i++ {
		res = append(res, &neuralnet.DenseLayer{
			InputCount:  14 * 14,
			OutputCount: 14 * 14,
		}, neuralnet.HyperbolicTangent{})
	}

	lastDepth := 1
	for i, outDepth := range []int{6, 15, 25, 20, 12} {
		if i > 0 {
			res = append(res, neuralnet.ReLU{})
		}
		res = append(res, &neuralnet.BorderLayer{
			InputWidth:   14,
			InputHeight:  14,
			InputDepth:   lastDepth,
			LeftBorder:   1,
			RightBorder:  1,
			TopBorder:    1,
			BottomBorder: 1,
		}, &neuralnet.ConvLayer{
			InputWidth:   16,
			InputHeight:  16,
			InputDepth:   lastDepth,
			Stride:       1,
			FilterCount:  outDepth,
			FilterWidth:  3,
			FilterHeight: 3,
		})
		lastDepth = outDepth
	}
	res = append(res, &neuralnet.UnstackLayer{
		InputWidth:    14,
		InputHeight:   14,
		InputDepth:    12,
		InverseStride: 2,
	}, neuralnet.Sigmoid{})
	res.Randomize()
	for _, layer := range res {
		if conv, ok := layer.(*neuralnet.ConvLayer); ok {
			for i := range conv.Biases.Vector {
				conv.Biases.Vector[i] = 1
			}
		}
	}
	return res
}

func createDiscriminator(samples sgd.SampleSet) neuralnet.Network {
	res := neuralnet.Network{
		&neuralnet.RescaleLayer{
			Bias:  -averageValue(samples),
			Scale: 1,
		},
	}
	width := 28
	height := 28
	depth := 3
	for i := 0; i < 3; i++ {
		conv := &neuralnet.ConvLayer{
			FilterCount:  10 + i*10,
			FilterWidth:  3,
			FilterHeight: 3,
			Stride:       1,
			InputWidth:   width,
			InputHeight:  height,
			InputDepth:   depth,
		}
		res = append(res, conv)
		res = append(res, neuralnet.ReLU{})
		max := &neuralnet.MaxPoolingLayer{
			InputWidth:  conv.OutputWidth(),
			InputHeight: conv.OutputHeight(),
			InputDepth:  conv.OutputDepth(),
			XSpan:       2,
			YSpan:       2,
		}
		res = append(res, max)
		width = max.OutputWidth()
		height = max.OutputHeight()
		depth = conv.OutputDepth()
	}
	res = append(res, &neuralnet.DenseLayer{
		InputCount:  width * height * depth,
		OutputCount: 100,
	})
	res = append(res, neuralnet.HyperbolicTangent{})
	res = append(res, &neuralnet.DenseLayer{
		InputCount:  100,
		OutputCount: 1,
	})
	res.Randomize()
	return res
}

func averageValue(s sgd.SampleSet) float64 {
	var sum float64
	var count float64
	for i := 0; i < s.Len(); i++ {
		vec := s.GetSample(i).(neuralnet.VectorSample).Input
		for _, x := range vec {
			sum += x
			count++
		}
	}
	return sum / count
}
