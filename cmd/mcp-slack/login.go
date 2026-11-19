package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rusq/slackauth"
)

// storedAuth est la session récupérée par « mcp-slack login », rangée en local.
// Rien de spécifique à une machine : n'importe quel utilisateur, n'importe quel
// workspace, n'importe quel poste.
type storedAuth struct {
	Workspace string `json:"workspace"`
	XOXC      string `json:"xoxc"`
	XOXD      string `json:"xoxd"`
}

// authConfigPath : <config utilisateur>/mcp-slack/auth.json (hors du repo).
func authConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mcp-slack", "auth.json"), nil
}

func loadStoredAuth() (*storedAuth, error) {
	p, err := authConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var a storedAuth
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func saveStoredAuth(a *storedAuth) (string, error) {
	p, err := authConfigPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		return "", err
	}
	return p, nil
}

// runLogin ouvre un navigateur, laisse l'utilisateur se connecter à son
// workspace Slack (SSO compris), puis récupère et range la session (le token
// xoxc et le cookie « d »). C'est le chemin agnostique : aucun token à copier
// à la main, aucune extension à installer.
func runLogin(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "usage : mcp-slack login <workspace>")
		fmt.Fprintln(os.Stderr, "  workspace = le sous-domaine, par exemple « etf1 » pour etf1.slack.com")
		os.Exit(2)
	}
	workspace := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Ouverture d'un navigateur pour se connecter à %s.slack.com...\n", workspace)
	fmt.Fprintln(os.Stderr, "Connecte-toi normalement (SSO compris). La fenêtre se ferme seule une fois la session récupérée.")

	client, err := slackauth.New(workspace, slackauth.WithBundledBrowser())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Impossible de préparer le navigateur : %v\n", err)
		os.Exit(1)
	}

	token, cookies, err := client.Manual(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connexion échouée : %v\n", err)
		os.Exit(1)
	}

	var d string
	for _, c := range cookies {
		if c.Name == "d" {
			d = c.Value
			break
		}
	}
	if token == "" || d == "" {
		fmt.Fprintln(os.Stderr, "Session incomplète : le token ou le cookie « d » manque.")
		os.Exit(1)
	}

	path, err := saveStoredAuth(&storedAuth{Workspace: workspace, XOXC: token, XOXD: d})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Impossible d'enregistrer la session : %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Session enregistrée dans %s\n", path)
	fmt.Fprintln(os.Stderr, "Le serveur mcp-slack l'utilisera automatiquement.")
}
