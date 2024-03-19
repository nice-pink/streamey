package audio

import (
	"errors"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/metricmanager"
)

// private bit validator

type PrivateBitValidator struct {
	active               bool
	verbose              bool
	metricManager        *metricmanager.MetricManager
	audioType            AudioType
	lastFrameDistance    uint64
	currentFrameCount    uint64
	parser               *Parser
	foundInitialDistance bool
}

func NewPrivateBitValidator(active bool, audioType AudioType, metricManager *metricmanager.MetricManager, verbose bool) *PrivateBitValidator {
	return &PrivateBitValidator{verbose: verbose, active: active, audioType: audioType, parser: NewParser(), metricManager: metricManager}
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
		if v.metricManager != nil {
			v.metricManager.IncParseAudioErrorCounter()
		}
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
		} else if v.verbose {
			log.Info("Found private bit.", i, v.currentFrameCount)
		}

		// validate distance
		if v.lastFrameDistance > 0 {
			if v.currentFrameCount != v.lastFrameDistance {
				if !v.foundInitialDistance {
					// skip first distance as it will usually be an error!
					v.foundInitialDistance = true
					log.Info("Skip initial distance.", v.currentFrameCount)
				} else {
					log.Error("Distances not equal. Current:", v.currentFrameCount, "!= Last:", v.lastFrameDistance, i, len(blockAudioInfo.Units))
					// update metric
					if v.metricManager != nil {
						v.metricManager.IncValidationErrorCounter()
					}
				}
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
	active        bool
	expectations  Expectations
	verbose       bool
	metricManager *metricmanager.MetricManager
	parser        *Parser
}

func NewEncodingValidator(active bool, expectations Expectations, metricManager *metricmanager.MetricManager, verbose bool) *EncodingValidator {
	return &EncodingValidator{expectations: expectations, verbose: verbose, active: active, metricManager: metricManager, parser: NewParser()}
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
	isValid := IsValid(v.expectations, *blockAudioInfo, v.metricManager)
	if !isValid && failEarly {
		return errors.New("validation failed")
	}

	// validate encodings
	for _, unit := range blockAudioInfo.Units {
		isValid = IsValidEncoding(v.expectations, unit.Encoding, v.metricManager)
		if !isValid && failEarly {
			return errors.New("validation failed")
		}
	}

	return nil
}

// general valiation

func IsValid(expectations Expectations, audioInfo AudioInfos, metricManager *metricmanager.MetricManager) bool {
	isValid := true
	if expectations.IsCBR {
		if expectations.IsCBR != audioInfo.IsCBR {
			log.Error("IsCBR not equal:", expectations.IsCBR, "!=", audioInfo.IsCBR)
			isValid = false
		}
	}

	// update metric
	if !isValid {
		if metricManager != nil {
			metricManager.IncValidationErrorCounter()
		}
	}

	return isValid
}

func IsValidEncoding(expectations Expectations, encoding Encoding, metricManager *metricmanager.MetricManager) bool {
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
		if metricManager != nil {
			metricManager.IncValidationErrorCounter()
		}
	}

	return isValid
}
