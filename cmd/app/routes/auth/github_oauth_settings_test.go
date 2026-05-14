package auth

import (
	"testing"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

func TestGithubOAuth2SettingsPreferProviderSnapshot(t *testing.T) {
	provider := pkgsettings.NewProvider()
	snapshot := settingtypes.DefaultApplicationSettings()
	snapshot.GithubOAuth2 = settingtypes.GithubOAuth2Settings{
		Enabled:      true,
		ClientID:     "persisted-client",
		ClientSecret: "persisted-secret",
		RedirectURL:  "http://localhost:3000/auth/github/callback",
		AuthURL:      settingtypes.DefaultGitHubAuthURL,
		TokenURL:     settingtypes.DefaultGitHubTokenURL,
		UserInfoURL:  settingtypes.DefaultGitHubUserInfoURL,
	}
	snapshot.Normalize()
	provider.Replace(snapshot, 9)

	controller := NewAuthController(&svc.ServiceContext{
		Config: &config.Config{
			GithubOAuth2: &options.GithubOAuth2Options{Enabled: false},
		},
		SettingsProvider: provider,
	})

	got := controller.githubOAuth2Settings()
	if !got.Enabled {
		t.Fatal("githubOAuth2Settings().Enabled = false, want true from provider")
	}
	if got.ClientID != "persisted-client" || got.ClientSecret != "persisted-secret" {
		t.Fatalf("githubOAuth2Settings() = %+v, want persisted provider credentials", got)
	}
}
