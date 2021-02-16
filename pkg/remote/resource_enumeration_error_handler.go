package remote

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/sirupsen/logrus"
)

type EnumerationAccessDeniedAlert struct {
	message string
}

func NewEnumerationAccessDeniedAlert(message string) EnumerationAccessDeniedAlert {
	return EnumerationAccessDeniedAlert{message}
}

func (e EnumerationAccessDeniedAlert) Message() string {
	return e.message
}

func (e EnumerationAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

func HandleResourceEnumerationError(err error, alertr *alerter.Alerter) error {
	listError, ok := err.(*remoteerror.ResourceEnumerationError)
	if !ok {
		return err
	}

	reqerr, ok := listError.RootCause().(awserr.RequestFailure)
	if !ok {
		return err
	}

	if reqerr.StatusCode() == 403 {
		message := fmt.Sprintf("Ignoring %s from drift calculation: Listing %s is forbidden.", listError.SupplierType(), listError.ListedTypeError())
		logrus.Debugf(message)
		alertr.SendAlert(listError.SupplierType(), NewEnumerationAccessDeniedAlert(message))
		return nil
	}

	return err
}
