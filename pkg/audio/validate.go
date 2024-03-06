package audio

import (
	"errors"

	"github.com/nice-pink/goutil/pkg/log"
)

type Validator struct {
	expectations Expectations
	verbose      bool
	active       bool
}

func NewValidator(active bool, expectations Expectations, verbose bool) *Validator {
	return &Validator{expectations: expectations, verbose: verbose, active: active}
}

func (v Validator) Validate(data []byte, failEarly bool) error {
	// bypass?
	if !v.active {
		return nil
	}

	// validate
	blockAudioInfo, err := ParseBlockwise(data, GetAudioTypeFromCodecName(v.expectations.Encoding.CodecName), true, v.verbose, false)
	if err != nil {
		log.Err(err, "Parsing error.")
		return err
	}

	if blockAudioInfo == nil {
		log.Error("No block audio data.")
		return nil
	}

	// validate audio info
	isValid := IsValid(v.expectations, *blockAudioInfo)
	if !isValid && failEarly {
		return errors.New("validation failed")
	}

	// validate encodings
	for _, unit := range blockAudioInfo.Units {
		isValid = IsValidEncoding(v.expectations, unit.Encoding)
		if !isValid && failEarly {
			return errors.New("validation failed")
		}
	}

	return nil
}

func IsValid(expectations Expectations, audioInfo AudioInfos) bool {
	isValid := true
	if expectations.IsCBR {
		if expectations.IsCBR != audioInfo.IsCBR {
			log.Error("IsCBR not equal:", expectations.IsCBR, "!=", audioInfo.IsCBR)
			isValid = false
		}
	}

	return isValid
}

func IsValidEncoding(expectations Expectations, encoding Encoding) bool {
	isValid := true
	if expectations.Encoding.Bitrate > 0 {
		if expectations.Encoding.Bitrate != encoding.Bitrate {
			log.Error("Bitrate not equal:", expectations.Encoding.Bitrate, "!=", encoding.Bitrate)
			isValid = false
		}
	}
	if expectations.Encoding.SampleRate > 0 {
		if expectations.Encoding.SampleRate != encoding.SampleRate {
			log.Error("SampleRate not equal:", expectations.Encoding.SampleRate, "!=", encoding.SampleRate)
			isValid = false
		}
	}
	if expectations.Encoding.FrameSize > 0 {
		if expectations.Encoding.FrameSize != encoding.FrameSize {
			log.Error("FrameSize not equal:", expectations.Encoding.FrameSize, "!=", encoding.FrameSize)
			isValid = false
		}
	}
	if expectations.Encoding.CodecName != "" {
		if expectations.Encoding.CodecName != encoding.CodecName {
			log.Error("CodecName not equal:", expectations.Encoding.CodecName, "!=", encoding.CodecName)
			isValid = false
		}
	}
	if expectations.Encoding.ContainerName != "" {
		if expectations.Encoding.ContainerName != encoding.ContainerName {
			log.Error("ContainerName not equal:", expectations.Encoding.ContainerName, "!=", encoding.ContainerName)
			isValid = false
		}
	}

	if expectations.Encoding.IsStereo != encoding.IsStereo {
		log.Error("IsStereo not equal:", expectations.Encoding.IsStereo, "!=", encoding.IsStereo)
		isValid = false
	}

	return isValid
}
