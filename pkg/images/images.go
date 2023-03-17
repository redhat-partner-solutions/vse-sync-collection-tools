package images

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)


var imageset = ImageSet{}

type ImageSet struct {
	NgenBaseImage string
	ready bool
}

const (
	DefaultNgenImage = "quay.io/jnunez/ngen_kpis:0.1"
)

func AddImage (imageName string) *ImageSet {
	if imageset.ready {
		return &imageset
	}
	if len(imageName) == 0 {
		log.Infof("use default image %v", DefaultNgenImage)
		imageName = DefaultNgenImage
	}
    imageset, err := NewImage(imageName)
    if err != nil {
    	log.Panic("Failed to collect images")
    }
    return imageset
}

func NewImage (imageName string) (*ImageSet,error) {
	log.Infof("creating new image from %v", imageName)
	if len(imageName) == 0 {
		log.Panic("must have at least one image to intialise imageset")
	}
	imageset.NgenBaseImage = imageName
	imageset.ready = true
	return &imageset , nil
}

func GetNgenBaseImage() (string, error) {
	if len(imageset.NgenBaseImage) == 0 {
		return "", fmt.Errorf("can't find a null NGEN base image name")
	}
	return imageset.NgenBaseImage, nil
}