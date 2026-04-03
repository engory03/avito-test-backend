package dummy

import "github.com/google/uuid"

// Фиксированные user_id для dummyLogin: один UUID на роль для стабильных тестов.
var (
	DummyAdminID = uuid.MustParse("01940000-0000-7000-8000-000000000001")
	DummyUserID  = uuid.MustParse("01940000-0000-7000-8000-000000000002")
)
