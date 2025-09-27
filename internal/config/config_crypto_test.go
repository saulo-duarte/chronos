package config_test

import (
	"os"
	"testing"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

const testKey = "01234567890123456789012345678901"

func TestInitCrypto(t *testing.T) {
	os.Setenv("CRYPTO_KEY", "chave_curta")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("InitCrypto deveria ter entrado em pânico com chave curta, mas não entrou.")
		}
	}()

	t.Run("ValidKey", func(t *testing.T) {
		os.Setenv("CRYPTO_KEY", testKey)

		config.InitCrypto()

	})
}

func TestEncryptDecrypt(t *testing.T) {
	os.Setenv("CRYPTO_KEY", testKey)
	config.InitCrypto()

	t.Run("SimpleText", func(t *testing.T) {
		plaintext := "dados de teste secretos"

		ciphertext, err := config.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt falhou com erro: %v", err)
		}

		decryptedtext, err := config.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt falhou com erro: %v", err)
		}

		if decryptedtext != plaintext {
			t.Errorf("O texto descriptografado ('%s') não corresponde ao original ('%s')",
				decryptedtext, plaintext)
		}

		ciphertext2, _ := config.Encrypt(plaintext)
		if ciphertext == ciphertext2 {
			t.Errorf("A criptografia não está sendo aleatória (nonce/IV). As cifras deveriam ser diferentes.")
		}
	})

	t.Run("EmptyText", func(t *testing.T) {
		plaintext := ""
		ciphertext, err := config.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt falhou com erro: %v", err)
		}
		decryptedtext, err := config.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt falhou com erro: %v", err)
		}
		if decryptedtext != plaintext {
			t.Errorf("O texto descriptografado vazio está incorreto: '%s'", decryptedtext)
		}
	})
}
