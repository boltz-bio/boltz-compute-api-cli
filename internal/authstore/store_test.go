package authstore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestRefreshTokenFallsBackToFileWhenKeyringUnavailable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)

	t.Cleanup(func() {
		ResetKeyring()
	})

	SetKeyringBackend(mockKeyringBackend{
		get: func(service, user string) (string, error) {
			return "", errors.New("keyring unavailable")
		},
		set: func(service, user, password string) error {
			return errors.New("keyring unavailable")
		},
		delete: func(service, user string) error {
			return keyring.ErrNotFound
		},
	})

	backend, err := SaveRefreshToken("refresh-token")
	require.NoError(t, err)
	require.Equal(t, "file", backend)

	token, loadedBackend, err := LoadRefreshToken()
	require.NoError(t, err)
	require.Equal(t, "refresh-token", token)
	require.Equal(t, "file", loadedBackend)

	require.NoError(t, ClearRefreshToken())

	token, loadedBackend, err = LoadRefreshToken()
	require.NoError(t, err)
	require.Empty(t, token)
	require.Empty(t, loadedBackend)
}

func TestLoadRefreshTokenWithPreferredBackendIgnoresStaleKeyringEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)

	t.Cleanup(func() {
		ResetKeyring()
	})

	staleKeyringToken := "stale-keyring-token"
	backend := mockKeyringBackend{
		get: func(service, user string) (string, error) {
			return staleKeyringToken, nil
		},
		set: func(service, user, password string) error {
			return errors.New("keyring unavailable")
		},
		delete: func(service, user string) error {
			staleKeyringToken = ""
			return nil
		},
	}
	SetKeyringBackend(backend)

	savedBackend, err := SaveRefreshToken("fresh-file-token")
	require.NoError(t, err)
	require.Equal(t, "file", savedBackend)

	token, loadedBackend, err := LoadRefreshTokenWithPreferredBackend("file")
	require.NoError(t, err)
	require.Equal(t, "fresh-file-token", token)
	require.Equal(t, "file", loadedBackend)

	token, loadedBackend, err = LoadRefreshToken()
	require.NoError(t, err)
	require.Equal(t, "fresh-file-token", token)
	require.Equal(t, "file", loadedBackend)
}

type mockKeyringBackend struct {
	get    func(service, key string) (string, error)
	set    func(service, key, value string) error
	delete func(service, key string) error
}

func (m mockKeyringBackend) Get(service, key string) (string, error) {
	if m.get != nil {
		return m.get(service, key)
	}
	return "", nil
}

func (m mockKeyringBackend) Set(service, key, value string) error {
	if m.set != nil {
		return m.set(service, key, value)
	}
	return nil
}

func (m mockKeyringBackend) Delete(service, key string) error {
	if m.delete != nil {
		return m.delete(service, key)
	}
	return nil
}
