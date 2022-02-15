package lock

import (
	"testing"
	"time"
)

var testConfig = map[Configuration]bool {
	Configuration {
		LockTimeout : 0,
		LockAttempts : 3,
		LockDuration : 30 * time.Second,
		TableName : "LOCK",
	} : false,
	Configuration {
		LockTimeout : 3 * time.Second,
		LockAttempts : 0,
		LockDuration : 30 * time.Second,
		TableName : "LOCK",
	} : false,
	Configuration {
		LockTimeout : 3 * time.Second,
		LockAttempts : 3,
		LockDuration : 0,
		TableName : "LOCK",
	} : false,
	Configuration {
		LockTimeout : 3 * time.Second,
		LockAttempts : 3,
		LockDuration : 30 * time.Second,
	} : false,
	Configuration {
		LockTimeout : 3 * time.Second,
		LockAttempts : 3,
		LockDuration : 30 * time.Second,
		TableName : "LOCK",
	} : true,
}

func TestConfigValidationNoTimeout(t *testing.T) {
	for configuration, expectedResult := range testConfig {
		err := configuration.validate()
		if !expectedResult && err == nil {
			t.Errorf("%v  expected to be invalid", configuration)
		}
	}

}