package mustacher

import (
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestMaxCorrelation(t *testing.T) {
	negativeCorrelations := []float64{0.9783704978644292, 0.9761565436397223, 0.9771920725385751}
	templateImg, err := readAssetImage("template.png")
	if err != nil {
		t.Fatal(err)
	}
	template := NewTemplate(templateImg)
	for i := 0; i < 3; i++ {
		image, err := readAssetImage("negative_" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			t.Error(err)
			continue
		}
		corr := template.MaxCorrelation(image)
		if math.Abs(corr-negativeCorrelations[i]) > 0.001 {
			t.Error("bad correlation", corr, "for negative", i)
		}
	}
	positiveImg, err := readAssetImage("positive.jpg")
	if corr := template.MaxCorrelation(positiveImg); math.Abs(corr-1) > 0.001 {
		t.Error("bad correlation", corr, "for positive")
	}
}

func BenchmarkMaxCorrelation(b *testing.B) {
	negatives := make([]*Image, 3)
	for i := range negatives {
		var err error
		negatives[i], err = readAssetImage("negative_" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
	templateImg, err := readAssetImage("template.png")
	if err != nil {
		b.Fatal(err)
	}
	template := NewTemplate(templateImg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, neg := range negatives {
			template.MaxCorrelation(neg)
		}
	}
}

func testAssetPath(filename string) string {
	_, sourceFilename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(sourceFilename), "test_assets", filename)
}

func readAssetImage(filename string) (*Image, error) {
	return ReadImageFile(testAssetPath(filename))
}
