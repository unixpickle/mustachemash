// Command train_placer trains a neural network to place
// mustaches in the correct position and orientation.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "image/jpeg"
	_ "image/png"

	"github.com/unixpickle/gans"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/weakai/neuralnet"
)

const (
	StepSize  = 0.00005
	BatchSize = 64
)

type Placement struct {
	ImageFile string
	CenterX   float64
	CenterY   float64
	Radius    float64
	Angle     float64
}

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s gan_in images placements.json net_out\n", os.Args[0])
		os.Exit(1)
	}

	log.Println("Making network...")
	network := makeNetwork(os.Args[1])

	log.Println("Loading samples...")
	samples := loadSamples(os.Args[2], os.Args[3])

	log.Println("Training network...")
	g := &neuralnet.BatchRGradienter{
		Learner:  network.BatchLearner(),
		CostFunc: neuralnet.MeanSquaredCost{},
	}
	var iteration int
	sgd.SGDMini(g, samples, StepSize, BatchSize, func(samples sgd.SampleSet) bool {
		cost := neuralnet.TotalCost(g.CostFunc, network, samples)
		log.Printf("iteration %d: cost=%f", iteration, cost)
		iteration++
		return true
	})

	log.Println("Saving network...")
	data, err := network.Serialize()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Serialize failed:", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(os.Args[4], data, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Save failed:", err)
		os.Exit(1)
	}
}

func makeNetwork(ganFile string) neuralnet.Network {
	data, err := ioutil.ReadFile(ganFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Read GAN failed:", err)
		os.Exit(1)
	}
	gan, err := gans.DeserializeFM(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Deserialize GAN failed:", err)
		os.Exit(1)
	}
	net := gan.Discriminator[:len(gan.Discriminator)-1]
	outLayer := &neuralnet.DenseLayer{
		InputCount:  100,
		OutputCount: 4,
	}
	outLayer.Randomize()
	net = append(net, outLayer)
	return net
}

func loadSamples(imageDir, placementFile string) sgd.SampleSet {
	var samples sgd.SliceSampleSet

	var placements []Placement
	placementData, err := ioutil.ReadFile(placementFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Read placements failed:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(placementData, &placements); err != nil {
		fmt.Fprintln(os.Stderr, "Decode placements failed:", err)
		os.Exit(1)
	}

	for _, placement := range placements {
		imgPath := filepath.Join(imageDir, placement.ImageFile)
		imgFile, err := os.Open(imgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Open image failed:", err)
			os.Exit(1)
		}
		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Decode image failed:", err)
			os.Exit(1)
		}
		tensor := imageTensor(img)
		outVec := []float64{
			placement.CenterX,
			placement.CenterY,
			placement.Radius,
			placement.Angle,
		}
		samples = append(samples, neuralnet.VectorSample{
			Input:  tensor.Data,
			Output: outVec,
		})
	}

	return samples
}

func imageTensor(img image.Image) *neuralnet.Tensor3 {
	res := neuralnet.NewTensor3(img.Bounds().Dx(), img.Bounds().Dy(), 3)
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x+img.Bounds().Min.X, y+img.Bounds().Min.Y).RGBA()
			res.Set(x, y, 0, float64(r)/0xffff)
			res.Set(x, y, 1, float64(g)/0xffff)
			res.Set(x, y, 2, float64(b)/0xffff)
		}
	}
	return res
}
