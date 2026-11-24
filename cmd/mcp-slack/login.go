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
	// XOXDS est la valeur du cookie « d-s » (simple horodatage).
	XOXDS string `json:"xoxd_s,omitempty"`
	// DBD est le cookie « dbd » (« v1-signed-… »), une signature d'appareil
	// signée par Slack. Slack l'exige désormais : sans lui, l'API renvoie 403
	// même avec « d » et « d-s » valides.
	DBD string `json:"dbd,omitempty"`
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

// applyStoredAuthEnv publie la session rangée dans les variables d'environnement
// lues par le serveur (xoxc, xoxd, d-s, dbd, workspace).
func applyStoredAuthEnv(a *storedAuth) {
	os.Setenv("SLACK_MCP_XOXC_TOKEN", a.XOXC)
	os.Setenv("SLACK_MCP_XOXD_TOKEN", a.XOXD)
	if a.XOXDS != "" {
		os.Setenv("SLACK_MCP_XOXD_S", a.XOXDS)
	}
	if a.DBD != "" {
		os.Setenv("SLACK_MCP_DBD", a.DBD)
	}
	if a.Workspace != "" {
		os.Setenv("SLACK_MCP_WORKSPACE", a.Workspace)
	}
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

// profileDir : le dossier de profil Chrome persistant, à côté de auth.json.
// La session Slack y reste connectée, ce qui permet au rafraîchissement
// silencieux de récupérer une session fraîche sans refaire le SSO.
func profileDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mcp-slack", "chrome-profile"), nil
}

// browserOption privilégie le navigateur installé (Chrome, Edge, Brave), à jour,
// plutôt que le Chromium embarqué figé que Slack refuse. L'embarqué ne sert
// qu'en dernier recours, et seulement en mode visible (il ne sait pas faire le
// SSO sans fenêtre).
func browserOption(verbose bool) []slackauth.Option {
	if browsers, err := slackauth.ListBrowsers(); err == nil && len(browsers) > 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "Navigateur utilisé : %s\n", browsers[0].Name)
		}
		return nil
	}
	if verbose {
		fmt.Fprintln(os.Stderr, "Aucun navigateur Chrome, Edge ou Brave trouvé : utilisation du navigateur embarqué.")
	}
	return []slackauth.Option{slackauth.WithBundledBrowser()}
}

// captureSession ouvre le workspace dans un profil Chrome persistant et récupère
// la session (token xoxc + cookies d, d-s, dbd), puis la range. Le navigateur
// est toujours visible : Slack ne démarre pas son client web sans fenêtre, donc
// le mode caché ne récupère rien. Mais tant que le profil est déjà connecté,
// aucun SSO n'est demandé (la fenêtre s'ouvre et se referme en ~15 s). « quiet »
// supprime les messages, pour le rafraîchissement automatique.
func captureSession(workspace string, quiet bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts := browserOption(!quiet)
	if dir, err := profileDir(); err == nil {
		if mkErr := os.MkdirAll(dir, 0o700); mkErr == nil {
			opts = append(opts, slackauth.WithUserDataDir(dir))
		}
	}

	client, err := slackauth.New(workspace, opts...)
	if err != nil {
		return fmt.Errorf("préparation du navigateur : %w", err)
	}
	token, cookies, err := client.Manual(ctx)
	if err != nil {
		return fmt.Errorf("récupération de la session : %w", err)
	}

	var d, ds, dbd string
	for _, c := range cookies {
		switch c.Name {
		case "d":
			d = c.Value
		case "d-s":
			ds = c.Value
		case "dbd":
			dbd = c.Value
		}
	}
	if token == "" || d == "" {
		return fmt.Errorf("session incomplète : token ou cookie « d » manquant")
	}
	if dbd == "" && !quiet {
		fmt.Fprintln(os.Stderr, "Avertissement : cookie « dbd » absent. Slack refuse désormais les appels sans lui (403).")
	}

	path, err := saveStoredAuth(&storedAuth{Workspace: workspace, XOXC: token, XOXD: d, XOXDS: ds, DBD: dbd})
	if err != nil {
		return fmt.Errorf("enregistrement de la session : %w", err)
	}
	if !quiet {
		fmt.Fprintf(os.Stderr, "Session enregistrée dans %s\n", path)
		fmt.Fprintln(os.Stderr, "Le serveur mcp-slack l'utilisera automatiquement.")
	}
	return nil
}

// runLogin : connexion interactive (fenêtre visible, SSO compris) dans le profil
// persistant. C'est le chemin agnostique : aucun token à copier, aucune
// extension à installer.
func runLogin(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "usage : mcp-slack login <workspace>")
		fmt.Fprintln(os.Stderr, "  workspace = le sous-domaine, par exemple « etf1 » pour etf1.slack.com")
		os.Exit(2)
	}
	workspace := args[0]
	fmt.Fprintf(os.Stderr, "Ouverture d'un navigateur pour se connecter à %s.slack.com...\n", workspace)
	fmt.Fprintln(os.Stderr, "Connecte-toi normalement (SSO compris). La fenêtre se ferme seule une fois la session récupérée.")
	if err := captureSession(workspace, false, 15*time.Minute); err != nil { // visible, verbeux
		fmt.Fprintf(os.Stderr, "Connexion échouée : %v\n", err)
		fmt.Fprintln(os.Stderr, "Si Slack répond « navigateur non pris en charge », installe Google Chrome puis relance : mcp-slack login "+workspace)
		os.Exit(1)
	}
}

// runRefresh : rafraîchissement silencieux (navigateur caché) depuis le profil
// persistant. Récupère une session fraîche sans SSO tant que le profil est
// connecté. Échoue vite sinon, pour que l'appelant sache qu'un SSO est requis.
func runRefresh(args []string) {
	workspace := ""
	if len(args) >= 1 {
		workspace = args[0]
	}
	if workspace == "" {
		if a, err := loadStoredAuth(); err == nil {
			workspace = a.Workspace
		}
	}
	if workspace == "" {
		fmt.Fprintln(os.Stderr, "usage : mcp-slack refresh <workspace> (aucune session connue)")
		os.Exit(2)
	}
	// Visible mais discret, délai large : la fenêtre s'ouvre ~15 s puis se ferme.
	// Tant que le profil est connecté, aucun SSO n'est demandé.
	if err := captureSession(workspace, true, 150*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Rafraîchissement impossible (SSO requis ? profil déconnecté ?) : %v\n", err)
		os.Exit(1)
	}
}
