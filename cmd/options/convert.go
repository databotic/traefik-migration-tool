package options

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type ConvertOptions struct {
	file     string
	output   string
	fileName string

	Input *os.File
	Out   *os.File
}

func NewConvertOptions() *ConvertOptions {
	o := &ConvertOptions{}
	return o
}

func (o *ConvertOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.file, "file", "f", "-", "filename or path to the resource to be converted.")
	fs.StringVarP(&o.output, "output", "o", "-", "output file")
}

func (o *ConvertOptions) Process() error {
	if o.file != "-" {
		stat, err := os.Stat(o.file)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(o.file, os.O_RDONLY, stat.Mode())
		if err != nil {
			return err
		}

		o.fileName = stat.Name()
		o.Input = file
	} else {
		o.Input = os.Stdin
	}

	if o.output != "-" {
		stat, err := os.Stat(o.output)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		if stat.IsDir() {
			o.output = filepath.Join(o.output, o.fileName)
		}

		open, err := os.OpenFile(o.output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		o.Out = open
	} else {
		o.Out = os.Stdout
	}

	return nil
}
