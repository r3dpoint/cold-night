package testutil

import (
	"reflect"
	"testing"
	"time"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected values to be different, but both were %v", message, actual)
	}
}

// AssertTrue checks if a condition is true
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true, got false", message)
	}
}

// AssertFalse checks if a condition is false
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false, got true", message)
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value != nil && !reflect.ValueOf(value).IsNil() {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertError checks if an error occurred
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected an error, got nil", message)
	}
}

// AssertNoError checks if no error occurred
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: expected no error, got %v", message, err)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, haystack, needle string, message string) {
	t.Helper()
	if !contains(haystack, needle) {
		t.Errorf("%s: expected '%s' to contain '%s'", message, haystack, needle)
	}
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, haystack, needle string, message string) {
	t.Helper()
	if contains(haystack, needle) {
		t.Errorf("%s: expected '%s' to not contain '%s'", message, haystack, needle)
	}
}

// AssertLengthEqual checks if a slice or array has the expected length
func AssertLengthEqual(t *testing.T, expected int, actual interface{}, message string) {
	t.Helper()
	v := reflect.ValueOf(actual)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		if v.Len() != expected {
			t.Errorf("%s: expected length %d, got %d", message, expected, v.Len())
		}
	default:
		t.Errorf("%s: cannot check length of type %T", message, actual)
	}
}

// AssertTimeBetween checks if a time is between two bounds
func AssertTimeBetween(t *testing.T, start, end, actual time.Time, message string) {
	t.Helper()
	if actual.Before(start) || actual.After(end) {
		t.Errorf("%s: expected time between %v and %v, got %v", message, start, end, actual)
	}
}

// AssertTimeEqual checks if two times are equal (within 1 second tolerance)
func AssertTimeEqual(t *testing.T, expected, actual time.Time, message string) {
	t.Helper()
	diff := actual.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("%s: expected time %v, got %v (diff: %v)", message, expected, actual, diff)
	}
}


// AssertPanic checks if a function panics
func AssertPanic(t *testing.T, fn func(), message string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, but function completed normally", message)
		}
	}()
	fn()
}

// AssertNoPanic checks if a function does not panic
func AssertNoPanic(t *testing.T, fn func(), message string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: expected no panic, but got: %v", message, r)
		}
	}()
	fn()
}

// Helper functions

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && 
		   (needle == "" || indexOfString(haystack, needle) >= 0)
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Benchmark helpers

// BenchmarkFunction runs a benchmark for a given function
func BenchmarkFunction(b *testing.B, fn func()) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// BenchmarkMemory runs a memory allocation benchmark
func BenchmarkMemory(b *testing.B, fn func()) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// Test setup helpers

// SetupTest provides common test setup
type TestSetup struct {
	EventStore *TestEventStore
	EventBus   *TestEventBus
	StartTime  time.Time
}

// NewTestSetup creates a new test setup with fresh dependencies
func NewTestSetup() *TestSetup {
	return &TestSetup{
		EventStore: NewTestEventStore(),
		EventBus:   NewTestEventBus(),
		StartTime:  time.Now(),
	}
}

// Cleanup performs any necessary cleanup after tests
func (ts *TestSetup) Cleanup() {
	// Reset state
	ts.EventStore = NewTestEventStore()
	ts.EventBus = NewTestEventBus()
}

// GetElapsedTime returns the time elapsed since test setup
func (ts *TestSetup) GetElapsedTime() time.Duration {
	return time.Since(ts.StartTime)
}

// Table test helpers

// TestCase represents a single test case for table-driven tests
type TestCase struct {
	Name           string
	Input          interface{}
	Expected       interface{}
	ExpectedError  bool
	ErrorMessage   string
	SetupFunc      func() interface{}
	CleanupFunc    func()
	SkipReason     string
}

// RunTableTests executes a series of table-driven tests
func RunTableTests(t *testing.T, testCases []TestCase, testFunc func(*testing.T, TestCase)) {
	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			if tc.SkipReason != "" {
				t.Skip(tc.SkipReason)
			}
			
			if tc.SetupFunc != nil {
				tc.SetupFunc()
			}
			
			if tc.CleanupFunc != nil {
				defer tc.CleanupFunc()
			}
			
			testFunc(t, tc)
		})
	}
}

// Mock helpers for external dependencies

// MockTime provides controllable time for testing
type MockTime struct {
	currentTime time.Time
}

func NewMockTime(initialTime time.Time) *MockTime {
	return &MockTime{currentTime: initialTime}
}

func (mt *MockTime) Now() time.Time {
	return mt.currentTime
}

func (mt *MockTime) Advance(duration time.Duration) {
	mt.currentTime = mt.currentTime.Add(duration)
}

func (mt *MockTime) Set(t time.Time) {
	mt.currentTime = t
}

// Performance testing helpers

// MeasureExecutionTime measures how long a function takes to execute
func MeasureExecutionTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// AssertExecutionTime checks if a function executes within expected time
func AssertExecutionTime(t *testing.T, maxDuration time.Duration, fn func(), message string) {
	t.Helper()
	duration := MeasureExecutionTime(fn)
	if duration > maxDuration {
		t.Errorf("%s: execution took %v, expected less than %v", message, duration, maxDuration)
	}
}