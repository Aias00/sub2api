package service

import (
	"context"
	"regexp"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Aias00/cloudbase/ent"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestRecordUserRegistrationEventPersistsRequestContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()

	svc := &AuthService{entClient: client}
	ctx := WithSignupGrantRiskInput(context.Background(), SignupGrantRiskInput{
		RemoteIP:          "203.0.113.10",
		UserAgent:         "Mozilla/5.0 Test",
		AcceptLanguage:    "zh-CN,zh;q=0.9",
		DeviceFingerprint: "device-123",
		ProviderType:      "google",
		ProviderSubject:   "sub-123",
		HeaderSnapshot: map[string]string{
			"User-Agent":      "Mozilla/5.0 Test",
			"Accept-Language": "zh-CN,zh;q=0.9",
		},
	})

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_registration_events")).
		WithArgs(
			int64(42),
			"user@example.com",
			"google",
			"google",
			"sub-123",
			"203.0.113.10",
			"Mozilla/5.0 Test",
			"zh-CN,zh;q=0.9",
			"device-123",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	svc.recordUserRegistrationEvent(ctx, &User{ID: 42, Email: "user@example.com", SignupSource: "google"}, "google")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
