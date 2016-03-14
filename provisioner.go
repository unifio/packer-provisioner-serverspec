package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// The command to run fabric
	Command string

	// Extra options to pass to the fabric command
	ExtraArguments []string `mapstructure:"extra_arguments"`

	FabricEnvVars []string `mapstructure:"fabric_env_vars"`

	// The main Fab file to execute.
	FabFile              string   `mapstructure:"fab_file"`
	FabTasks             string   `mapstructure:"fab_tasks"`
	User                 string   `mapstructure:"user"`
	LocalPort            string   `mapstructure:"local_port"`
	SSHHostKeyFile       string   `mapstructure:"ssh_host_key_file"`
	SSHAuthorizedKeyFile string   `mapstructure:"ssh_authorized_key_file"`
}

type Provisioner struct {
	config        Config
	adapter       *adapter
	done          chan struct{}
	fabVersion    string
	fabMajVersion uint
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	p.done = make(chan struct{})

	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	if p.config.Command == "" {
		p.config.Command = "fab"
	}

	var errs *packer.MultiError
	err = validateFileConfig(p.config.FabFile, "fab_file", true)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	// Check that the authorized key file exists
	if len(p.config.SSHAuthorizedKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHAuthorizedKeyFile, "ssh_authorized_key_file", true)
		if err != nil {
			log.Println(p.config.SSHAuthorizedKeyFile, "does not exist")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}
	if len(p.config.SSHHostKeyFile) > 0 {
		err = validateFileConfig(p.config.SSHHostKeyFile, "ssh_host_key_file", true)
		if err != nil {
			log.Println(p.config.SSHHostKeyFile, "does not exist")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if len(p.config.LocalPort) > 0 {
		if _, err := strconv.ParseUint(p.config.LocalPort, 10, 16); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("local_port: %s must be a valid port", p.config.LocalPort))
		}
	} else {
		p.config.LocalPort = "0"
	}

	err = p.getVersion()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if p.config.User == "" {
		p.config.User = os.Getenv("USER")
	}
	if p.config.User == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("user: could not determine current user from environment."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) getVersion() error {
	out, err := exec.Command(p.config.Command, "--version").Output()
	if err != nil {
		return err
	}

	versionRe := regexp.MustCompile(`Fabric (\d+\.\d+[.\d+]*)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return fmt.Errorf(
			"Could not find %s version in output:\n%s", p.config.Command, string(out))
	}

	version := matches[1]
	log.Printf("%s version: %s", p.config.Command, version)
	p.fabVersion = version

	majVer, err := strconv.ParseUint(strings.Split(version, ".")[0], 10, 0)
	if err != nil {
		return fmt.Errorf("Could not parse major version from \"%s\".", version)
	}
	p.fabMajVersion = uint(majVer)

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Fabric...")

	k, err := newUserKey(p.config.SSHAuthorizedKeyFile)
	if err != nil {
		return err
	}

	hostSigner, err := newSigner(p.config.SSHHostKeyFile)
	// Remove the private key file
	if len(k.privKeyFile) > 0 {
		defer os.Remove(k.privKeyFile)
	}

	keyChecker := ssh.CertChecker{
		UserKeyFallback: func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if user := conn.User(); user != p.config.User {
				ui.Say(fmt.Sprintf("%s is not a valid user", user))
				return nil, errors.New("authentication failed")
			}

			if !bytes.Equal(k.Marshal(), pubKey.Marshal()) {
				ui.Say("unauthorized key")
				return nil, errors.New("authentication failed")
			}

			return nil, nil
		},
	}

	config := &ssh.ServerConfig{
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			ui.Say(fmt.Sprintf("authentication attempt from %s to %s as %s using %s", conn.RemoteAddr(), conn.LocalAddr(), conn.User(), method))
		},
		PublicKeyCallback: keyChecker.Authenticate,
		//NoClientAuth:      true,
	}

	config.AddHostKey(hostSigner)

	localListener, err := func() (net.Listener, error) {
		port, err := strconv.ParseUint(p.config.LocalPort, 10, 16)
		if err != nil {
			return nil, err
		}

		tries := 1
		if port != 0 {
			tries = 10
		}
		for i := 0; i < tries; i++ {
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			port++
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			_, p.config.LocalPort, err = net.SplitHostPort(l.Addr().String())
			if err != nil {
				ui.Say(err.Error())
				continue
			}
			return l, nil
		}
		return nil, errors.New("Error setting up SSH proxy connection")
	}()

	if err != nil {
		return err
	}

	ui = newUi(ui)
	p.adapter = newAdapter(p.done, localListener, config, "", ui, comm)

	defer func() {
		ui.Say("shutting down the SSH proxy")
		close(p.done)
		p.adapter.Shutdown()
	}()

	go p.adapter.Serve()

	if err := p.executeFabric(ui, comm, k.privKeyFile, !hostSigner.generated); err != nil {
		return fmt.Errorf("Error executing Fabric: %s", err)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	if p.done != nil {
		close(p.done)
	}
	if p.adapter != nil {
		p.adapter.Shutdown()
	}
	os.Exit(0)
}

func (p *Provisioner) executeFabric(ui packer.Ui, comm packer.Communicator, privKeyFile string, checkHostKey bool) error {
  fabfile, _ := filepath.Abs(p.config.FabFile)
	hoststring := fmt.Sprintf("%s@127.0.0.1:%s",
		p.config.User, p.config.LocalPort)
	var envvars []string

	args := []string{"-f", fabfile, "-H", hoststring}
	if len(privKeyFile) > 0 {
		args = append(args, "-i", privKeyFile)
	}
	if !checkHostKey {
		args = append(args, "--disable-known-hosts")
	}
	args = append(args, p.config.ExtraArguments...)
	args = append(args, p.config.FabTasks)

	if len(p.config.FabricEnvVars) > 0 {
		envvars = append(envvars, p.config.FabricEnvVars...)
	}

	cmd := exec.Command(p.config.Command, args...)

	cmd.Env = os.Environ()
	if len(envvars) > 0 {
		cmd.Env = append(cmd.Env, envvars...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	repeat := func(r io.ReadCloser) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			ui.Message(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			ui.Error(err.Error())
		}
		wg.Done()
	}
	wg.Add(2)
	go repeat(stdout)
	go repeat(stderr)

	ui.Say(fmt.Sprintf("Executing Fabric: %s", strings.Join(cmd.Args, " ")))
	cmd.Start()
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Non-zero exit status: %s", err)
	}

	return nil
}

func validateFileConfig(name string, config string, req bool) error {
	if req {
		if name == "" {
			return fmt.Errorf("%s must be specified.", config)
		}
	}
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, name, err)
	} else if info.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, name)
	}
	return nil
}

type userKey struct {
	ssh.PublicKey
	privKeyFile string
}

func newUserKey(pubKeyFile string) (*userKey, error) {
	userKey := new(userKey)
	if len(pubKeyFile) > 0 {
		pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
		if err != nil {
			return nil, errors.New("Failed to read public key")
		}
		userKey.PublicKey, _, _, _, err = ssh.ParseAuthorizedKey(pubKeyBytes)
		if err != nil {
			return nil, errors.New("Failed to parse authorized key")
		}

		return userKey, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("Failed to generate key pair")
	}
	userKey.PublicKey, err = ssh.NewPublicKey(key.Public())
	if err != nil {
		return nil, errors.New("Failed to extract public key from generated key pair")
	}

	// To support Fabric calling back to us we need to write
	// this file down
	privateKeyDer := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	tf, err := ioutil.TempFile("", "fabric-key")
	if err != nil {
		return nil, errors.New("failed to create temp file for generated key")
	}
	_, err = tf.Write(pem.EncodeToMemory(&privateKeyBlock))
	if err != nil {
		return nil, errors.New("failed to write private key to temp file")
	}

	err = tf.Close()
	if err != nil {
		return nil, errors.New("failed to close private key temp file")
	}
	userKey.privKeyFile = tf.Name()

	return userKey, nil
}

type signer struct {
	ssh.Signer
	generated bool
}

func newSigner(privKeyFile string) (*signer, error) {
	signer := new(signer)

	if len(privKeyFile) > 0 {
		privateBytes, err := ioutil.ReadFile(privKeyFile)
		if err != nil {
			return nil, errors.New("Failed to load private host key")
		}

		signer.Signer, err = ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			return nil, errors.New("Failed to parse private host key")
		}

		return signer, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("Failed to generate server key pair")
	}

	signer.Signer, err = ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, errors.New("Failed to extract private key from generated key pair")
	}
	signer.generated = true

	return signer, nil
}

// Ui provides concurrency-safe access to packer.Ui.
type Ui struct {
	sem chan int
	ui  packer.Ui
}

func newUi(ui packer.Ui) packer.Ui {
	return &Ui{sem: make(chan int, 1), ui: ui}
}

func (ui *Ui) Ask(s string) (string, error) {
	ui.sem <- 1
	ret, err := ui.ui.Ask(s)
	<-ui.sem

	return ret, err
}

func (ui *Ui) Say(s string) {
	ui.sem <- 1
	ui.ui.Say(s)
	<-ui.sem
}

func (ui *Ui) Message(s string) {
	ui.sem <- 1
	ui.ui.Message(s)
	<-ui.sem
}

func (ui *Ui) Error(s string) {
	ui.sem <- 1
	ui.ui.Error(s)
	<-ui.sem
}

func (ui *Ui) Machine(t string, args ...string) {
	ui.sem <- 1
	ui.ui.Machine(t, args...)
	<-ui.sem
}
