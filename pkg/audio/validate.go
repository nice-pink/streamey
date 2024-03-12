package audio

import (
	"errors"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/metricmanager"
)

// private bit validator

type PrivateBitValidator struct {
	active            bool
	verbose           bool
	metrics           bool
	audioType         AudioType
	lastFrameDistance uint64
	currentFrameCount uint64
	parser            *Parser
}

func NewPrivateBitValidator(active bool, audioType AudioType, metrics bool, verbose bool) *PrivateBitValidator {
	return &PrivateBitValidator{verbose: verbose, active: active, audioType: audioType, parser: NewParser(), metrics: metrics}
}

func (v *PrivateBitValidator) Validate(data []byte, failEarly bool) error {
	// bypass?
	if !v.active {
		return nil
	}

	// validate
	blockAudioInfo, err := v.parser.ParseBlockwise(data, v.audioType, true, v.verbose, false)
	if err != nil {
		log.Err(err, "Parsing error.")
		metricmanager.IncParseAudioErrorCounter()
		return err
	}

	if blockAudioInfo == nil {
		// log.Error("No block audio data.")
		return nil
	}

	// validate encodings
	for i, unit := range blockAudioInfo.Units {
		if !unit.IsPrivate {
			v.currentFrameCount++
			continue
		} else {
			log.Info("Found private bit.", i, v.currentFrameCount)
		}

		// validate distance
		if v.lastFrameDistance > 0 {
			if v.currentFrameCount != v.lastFrameDistance {
				log.Error("Distances not equal. Current:", v.currentFrameCount, "!= Last:", v.lastFrameDistance, i, len(blockAudioInfo.Units))

				// update metric
				metricmanager.IncValidationErrorCounter()
			} // else {
			// 	log.Info("Distance between private bits:", v.lastFrameDistance)
			// }
		}

		// reset
		v.lastFrameDistance = v.currentFrameCount
		v.currentFrameCount = 0
	}

	return nil
}

// encoding validator

type EncodingValidator struct {
	active       bool
	expectations Expectations
	verbose      bool
	metrics      bool
	parser       *Parser
}

func NewEncodingValidator(active bool, expectations Expectations, metrics bool, verbose bool) *EncodingValidator {
	return &EncodingValidator{expectations: expectations, verbose: verbose, active: active, metrics: metrics, parser: NewParser()}
}

func (v *EncodingValidator) Validate(data []byte, failEarly bool) error {
	// bypass?
	if !v.active {
		return nil
	}

	// validate
	blockAudioInfo, err := v.parser.ParseBlockwise(data, GetAudioTypeFromCodecName(v.expectations.Encoding.CodecName), true, v.verbose, false)
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

// general valiation

func IsValid(expectations Expectations, audioInfo AudioInfos) bool {
	isValid := true
	if expectations.IsCBR {
		if expectations.IsCBR != audioInfo.IsCBR {
			log.Error("IsCBR not equal:", expectations.IsCBR, "!=", audioInfo.IsCBR)
			isValid = false
		}
	}

	// update metric
	if !isValid {
		metricmanager.IncValidationErrorCounter()
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

	// update metric
	if !isValid {
		metricmanager.IncValidationErrorCounter()
	}

	return isValid
}
