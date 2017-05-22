package util

import (
	"testing"
	"reflect"
)

func TestStringToMap(t *testing.T) {
	var expected = map[string]string{"key1": "value1", "key2": "value2"}

	actual := StringToMap("key1=value1,key2=value2")

	assertEqual(t, expected, actual)
}

func TestMapToSortedList(t *testing.T) {
	var input = map[string]string{"ckey1": "value1", "zkey2":"value2", "bkey3": "value3"}

	var expected = []string{"bkey3=value3", "ckey1=value1", "zkey2=value2"}

	output := MapToSortedList(input)

	assertEqual(t, expected, output);
}

func TestMergeMapsAddsValues(t *testing.T) {
	var expected = map[string]string{"ckey1": "value1", "zkey2":"value2", "bkey3": "value3"}
	var input = map[string]string{"ckey1": "value1", "bkey3": "value3"}
	var defaultValues = map[string]string{"zkey2":"value2"}

	output := MergeMaps(input, defaultValues)

	assertEqual(t, expected, output)
}

func TestMergeMapsDoesNotOverrideInputValues(t *testing.T) {
	var expected = map[string]string{"ckey1": "value1", "zkey2":"value2", "bkey3": "value3"}
	var input = map[string]string{"ckey1": "value1", "bkey3": "value3", "zkey2":"value2"}
	var defaultValues = map[string]string{"zkey2":"value7"}

	output := MergeMaps(input, defaultValues)

	assertEqual(t, expected, output)
}

func TestUniqueListOfKeys(t *testing.T) {
	var map1 = map[string]string{"ckey1": "value1", "zkey2":"value2", "bkey3": "value3"}
	var map2 = map[string]string{"ckey1": "value1", "zkey2":"value2", "dkey4": "value4"}
	var map3 = map[string]string{"ckey5": "value5", "zkey2":"value2", "dkey6": "value6"}

	expected := []string{"bkey3", "ckey1", "ckey5", "dkey4", "dkey6", "zkey2"}

	output := MapKeys(map1, map2, map3)

	assertEqual(t, expected, output)
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}