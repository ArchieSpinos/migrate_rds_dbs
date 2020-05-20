package awsresources

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateSession(region string, profile string) (*session.Session, *errors.DBErr) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		return nil, errors.NewBadRequestError(fmt.Sprintf("failed to setup aws session: %s", err.Error()))
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		return nil, errors.NewBadRequestError(fmt.Sprintf("failed to retrieve aws credentials: %s", err.Error()))
	}
	return sess, nil
}
