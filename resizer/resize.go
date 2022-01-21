package resizer

import (
	"github.com/sirupsen/logrus"
	"github.com/tokubai/kinu/engine"
	"github.com/tokubai/kinu/logger"
)

func Resize(image []byte, option *ResizeOption) (result *ResizeResult) {
	calculator, err := NewCoodinatesCalculator(option)
	if err != nil {
		return &ResizeResult{err: logger.ErrorDebug(err)}
	}

	engine, err := engine.New(image)
	if err != nil {
		return &ResizeResult{err: logger.ErrorDebug(err)}
	}

	var coodinates *Coodinates
	if option.HasSizeHint() && !option.NeedsManualCrop {
		calculator.SetImageSize(option.SizeHintWidth, option.SizeHintHeight)
		coodinates = calculator.Calc(option)
		engine.SetSizeHint(coodinates.ResizeWidth, coodinates.ResizeHeight)
		logger.WithFields(logrus.Fields{
			"width_size_hint":  coodinates.ResizeWidth,
			"height_size_hint": coodinates.ResizeHeight,
		}).Debug("size hint")
	} else {
		logger.Debug("not set size hint")
	}

	err = engine.Open()
	if err != nil {
		return &ResizeResult{err: logger.ErrorDebug(err)}
	}

	defer engine.Close()

	if coodinates == nil {
		calculator.SetImageSize(engine.GetImageWidth(), engine.GetImageHeight())
		coodinates = calculator.Calc(option)
	}

	if option.NeedsManualCrop {
		// crop first then resize for manual cropping.
		err = engine.Crop(coodinates.CropWidth, coodinates.CropHeight, coodinates.WidthOffset, coodinates.HeightOffset)
		if err != nil {
			return &ResizeResult{err: logger.ErrorDebug(err)}
		}

		err = engine.Resize(coodinates.ResizeWidth, coodinates.ResizeHeight)
		if err != nil {
			return &ResizeResult{err: logger.ErrorDebug(err)}
		}
	} else {
		err = engine.Resize(coodinates.ResizeWidth, coodinates.ResizeHeight)
		if err != nil {
			return &ResizeResult{err: logger.ErrorDebug(err)}
		}

		if coodinates.CanCrop() {
			err = engine.Crop(coodinates.CropWidth, coodinates.CropHeight, coodinates.WidthOffset, coodinates.HeightOffset)
			if err != nil {
				return &ResizeResult{err: logger.ErrorDebug(err)}
			}
		}
	}

	if option.HasAlphaChannel() && option.NeedsRemoveAlpha() {
		logger.Debug("removing alpha channel")
		err = engine.RemoveAlpha()
		if err != nil {
			return &ResizeResult{err: logger.ErrorDebug(err)}
		}
	}

	if len(option.Format) != 0 {
		engine.SetFormat(option.Format)
	}

	if option.Quality != 0 {
		engine.SetCompressionQuality(option.Quality)
	}

	resultImage, err := engine.Generate()
	return &ResizeResult{image: resultImage, err: err}
}
