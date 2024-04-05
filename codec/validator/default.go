package validator

import "errors"

var (
	ErrNotNull = "notnull validation failed for field %s"

	ErrMin = "min value validation failed for field %s"

	ErrMax = "max value validation failed for field %s"

	ErrExclusiveMin = "exclusive min validation failed for field %s"

	ErrExclusiveMax = "exclusive max validation failed for field %s"

	ErrMultipleOf = "multipleOf validation failed for field %s"

	ErrMinLength = "min-length validation failed for field %s"

	ErrMaxLength = "max-length validation failed for field %s"

	ErrPattern = "pattern validation failed for field %s"

	ErrEnums = "enum validation failed for field %s"

	ErrBadConstraint = "invalid constraint %s with value '%s' for field %s"

	ErrInvalidValidationForField = "invalid validation applied to the field %s"

	ErrNotSupported = errors.New("unsupported constraint on type")

	ErrConversionFailed = errors.New("conversion failed")
)
