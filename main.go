package main

import (
			"flag"
      "fmt"
			"os"
			"path/filepath"

      "github.com/nats-io/jwt/v2"
      "github.com/nats-io/nkeys"
      "gopkg.in/yaml.v2"
)

func main() {
	outputDir := flag.String("output-dir", "", "Output directory");
	flag.Parse()
	if *outputDir == "" {
		fmt.Println("--output-dir required")
		os.Exit(1)
	}

  operatorKP,err := nkeys.CreateOperator()
	if err != nil {
		panic(err)
	}
  operatorSeed, err := operatorKP.Seed()
	if err != nil {
		panic(err)
	}
  operatorSeedString := string(operatorSeed)
  operatorPub, err := operatorKP.PublicKey()
	if err != nil {
		panic(err)
	}

  sysAccountKP, err := nkeys.CreateAccount()
	if err != nil {
		panic(err)
	}
  sysAccountPub, err := sysAccountKP.PublicKey()
	if err != nil {
		panic(err)
	}
  sysAccountPubString := string(sysAccountPub)
	sysAccountSeed, err := sysAccountKP.Seed()
	if err != nil {
		panic(err)
	}
	sysAccountSeedString := string(sysAccountSeed)
  sysAccountClaims := jwt.NewAccountClaims(sysAccountPub)
  sysAccountClaims.Name = "SYS"
  sysAccountJWT, err := sysAccountClaims.Encode(operatorKP)
	if err != nil {
		panic(err)
	}

  operatorClaims := jwt.NewOperatorClaims(operatorPub)
  operatorClaims.Name = "coflow-system"
	operatorClaims.SystemAccount = sysAccountPubString
	operatorClaims.AccountServerURL = "nats://0.0.0.0:4222"
  operatorJWT, err := operatorClaims.Encode(operatorKP)
	if err != nil {
		panic(err)
	}

	sysUserKP, err := nkeys.CreateUser()
	if err != nil {
		panic(err)
	}
	sysUserPub, err := sysUserKP.PublicKey()
	if err != nil {
		panic(err)
	}
	sysUserSeed, err := sysUserKP.Seed()
	if err != nil {
		panic(err)
	}
	sysUserSeedString := string(sysUserSeed)
	sysUserClaims := jwt.NewUserClaims(sysUserPub)
	sysUserClaims.Name = "sys"
	sysUserClaims.IssuerAccount = sysAccountPubString
	sysUserJWT, err := sysUserClaims.Encode(sysAccountKP)
	if err != nil {
		panic(err)
	}

  resolverData := Config{
    Config: ConfigWrapper{
      Merge: Merge{
        Operator:         operatorJWT,
        SystemAccount:    sysAccountPubString,
        ResolverPreload:  map[string]string{
          sysAccountPubString: sysAccountJWT,
        },
      },
    },
	}

	// Make output directory
	err = os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		os.Exit(1)
	}

	// Write resolver.yaml
  resolverYaml, err := yaml.Marshal(&resolverData)
	if err != nil {
		panic(err)
	}
	resolverPath := filepath.Join(*outputDir, "resolver.yaml")
	err = os.WriteFile(resolverPath, []byte(resolverYaml), 0644)
	if err != nil {
		fmt.Println("Error write resolver.yaml")
		os.Exit(1)
	}

	// Write operator JWT
	operatorJWTPath := filepath.Join(*outputDir, "operator.jwt")
	err = os.WriteFile(operatorJWTPath, []byte(operatorJWT), 0644)
	if (err != nil) {
		fmt.Println("Error writing operator.jwt", err)
		os.Exit(1)
	}

	// Write operator Seed
	operatorSeedPath := filepath.Join(*outputDir, "operator.seed")
	err = os.WriteFile(operatorSeedPath, []byte(operatorSeedString), 0644)
	if (err != nil) {
		fmt.Println("Error writing operator seed")
		os.Exit(1)
	}

	// Write account seed
	accountSeedPath := filepath.Join(*outputDir, "account.seed")
	err = os.WriteFile(accountSeedPath, []byte(sysAccountSeedString), 0644)
	if (err != nil) {
		fmt.Println("Error writing account seed")
		os.Exit(1)
	}

	// Write sys.creds
	creds := GenerateCredsContent(sysUserJWT, sysUserSeedString)
	credsPath := filepath.Join(*outputDir, "sys.creds")
	err = os.WriteFile(credsPath, []byte(creds), 0644)
	if err != nil {
		fmt.Println("Error writing sys.creds")
		os.Exit(1)
	}

	fmt.Printf("Secrets written to %s\n", *outputDir)
}

func GenerateCredsContent(jwt string, seed string) string {
	credsContent := fmt.Sprintf(`-----BEGIN NATS USER JWT-----
%s
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.
-----BEGIN USER NKEY SEED-----
%s
------END USER NKEY SEED------
`, jwt, seed)
	return credsContent
}


type Config struct {
  Config ConfigWrapper `yaml:"config"`
}

type ConfigWrapper struct {
  Merge Merge `yaml:"merge"`
}

type Merge struct {
  Operator        string `yaml:"operator"`
  SystemAccount   string `yaml:"system_account"`
  ResolverPreload map[string]string `yaml:"resolver_preload"`
}

