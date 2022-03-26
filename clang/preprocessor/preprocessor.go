package preprocessor

import (
	"os"
	"os/exec"
)

// -----------------------------------------------------------------------------

type Config struct {
	Compiler    string // default: clang
	PPFlag      string // default: -E
	IncludeDirs []string
	Defines     []string
	Flags       []string
}

func Do(infile, outfile string, conf *Config) (err error) {
	if conf == nil {
		conf = new(Config)
	}
	compiler := conf.Compiler
	if compiler == "" {
		compiler = "clang"
	}
	ppflag := conf.PPFlag
	if ppflag == "" {
		ppflag = "-E"
	}
	n := 4 + len(conf.Flags) + len(conf.IncludeDirs) + len(conf.Defines)
	args := make([]string, 3, n)
	args[0] = ppflag
	args[1], args[2] = "-o", outfile
	args = append(args, conf.Flags...)
	for _, def := range conf.Defines {
		args = append(args, "-D"+def)
	}
	for _, inc := range conf.IncludeDirs {
		args = append(args, "-I"+inc)
	}
	args = append(args, infile)
	cmd := exec.Command(compiler, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// -----------------------------------------------------------------------------
