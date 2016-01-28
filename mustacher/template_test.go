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
	if err != nil {
		t.Fatal(err)
	}
	if corr := template.MaxCorrelation(positiveImg); math.Abs(corr-1) > 0.001 {
		t.Error("bad correlation", corr, "for positive")
	}
}

func TestCorrelations(t *testing.T) {
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
		matches := template.Correlations(image, 0.98)
		if len(matches) > 0 {
			t.Error("got false matches for negative", i)
		}
	}
	positiveImg, err := readAssetImage("positive.jpg")
	if err != nil {
		t.Fatal(err)
	}
	matches := template.Correlations(positiveImg, 0.98)
	matches = matches.NonOverlappingSet()
	if len(matches) != 1 {
		t.Error("invalid number of matches for positive:", len(matches))
	} else if matches[0].X != 201 || matches[0].Y != 593 {
		t.Error("invalid match for positive:", matches[0])
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

func BenchmarkCorrelations(b *testing.B) {
	templateImg, err := readAssetImage("template.png")
	if err != nil {
		b.Fatal(err)
	}
	template := NewTemplate(templateImg)
	positiveImg, err := readAssetImage("positive.jpg")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		template.Correlations(positiveImg, 0.98)
	}
}

func testAssetPath(filename string) string {
	_, sourceFilename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(sourceFilename), "test_assets", filename)
}

func readAssetImage(filename string) (*Image, error) {
	return ReadImageFile(testAssetPath(filename))
}
