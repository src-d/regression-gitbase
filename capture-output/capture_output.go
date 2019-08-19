package capture_output

import (
	"io/ioutil"
	"os"
	"time"
)

// Capture returns stdout and stderr output of the function as a string
func Capture(f func(), delay time.Duration) string {
	file, err := ioutil.TempFile("", "prefix")
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = file
	os.Stderr = file

	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		os.RemoveAll(file.Name())
	}()

	f()
	time.Sleep(delay)
	if err := file.Close(); err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}

	return string(data)
}
