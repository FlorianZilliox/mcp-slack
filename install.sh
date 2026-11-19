#!/bin/sh
# Installe mcp-slack : telecharge le programme pre-compile pour ta machine,
# l'ajoute a Claude Code, et ouvre une fenetre pour te connecter a Slack.
# Fonctionne sur macOS et Linux. Aucune installation prealable requise.
set -eu

REPO="florianzilliox/mcp-slack"
BIN_DIR="$HOME/.local/bin"
BIN="$BIN_DIR/mcp-slack"

# Affiche la marche a suivre quand l'ajout automatique a Claude n'est pas possible.
print_manual_config() {
  cat <<EOF

Ajoute mcp-slack a la main : ouvre (ou cree) le fichier ~/.claude.json et place
mcp-slack dans la section "mcpServers" :

  {
    "mcpServers": {
      "mcp-slack": {
        "command": "$BIN"
      }
    }
  }

Avec l'extension VS Code : ce meme fichier ~/.claude.json est utilise. Tu peux
aussi ouvrir le terminal integre de VS Code et lancer :

  claude mcp add mcp-slack --scope user -- "$BIN"

Recharge ensuite la fenetre VS Code.
EOF
}

echo "Installation de mcp-slack..."

# 1. Detecter le systeme d'exploitation.
os=$(uname -s)
case "$os" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)
    echo "Erreur : systeme non reconnu ($os)."
    echo "mcp-slack s'installe sur macOS et Linux. Sur Windows, demande de l'aide a la personne qui t'a partage cet outil."
    exit 1
    ;;
esac

# 2. Detecter l'architecture du processeur.
arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Erreur : type de processeur non reconnu ($arch)."
    echo "Previens la personne qui t'a partage cet outil : il faut un programme adapte a ta machine."
    exit 1
    ;;
esac

ASSET="mcp-slack-$OS-$ARCH"
URL="https://github.com/$REPO/releases/latest/download/$ASSET"

# 3. Telecharger le programme.
echo "Telechargement du programme pour ta machine ($OS $ARCH)..."
tmp=$(mktemp)
if command -v curl >/dev/null 2>&1; then
  if ! curl -fL --progress-bar -o "$tmp" "$URL"; then
    rm -f "$tmp"
    echo "Erreur : le telechargement a echoue."
    echo "Verifie ta connexion internet, puis relance la commande d'installation."
    echo "Si le probleme persiste, previens la personne qui t'a partage cet outil : il manque peut-etre une version publiee pour ta machine."
    exit 1
  fi
elif command -v wget >/dev/null 2>&1; then
  if ! wget -O "$tmp" "$URL"; then
    rm -f "$tmp"
    echo "Erreur : le telechargement a echoue."
    echo "Verifie ta connexion internet, puis relance la commande d'installation."
    exit 1
  fi
else
  rm -f "$tmp"
  echo "Erreur : ni curl ni wget ne sont disponibles sur cette machine."
  echo "Installe l'un des deux, puis relance la commande d'installation."
  exit 1
fi

# Un fichier vide signale une release absente ou une page d'erreur recue a la place du programme.
if [ ! -s "$tmp" ]; then
  rm -f "$tmp"
  echo "Erreur : le fichier telecharge est vide."
  echo "La version publiee est peut-etre indisponible pour ta machine. Previens la personne qui t'a partage cet outil."
  exit 1
fi

# 4. Installer le programme dans ~/.local/bin.
mkdir -p "$BIN_DIR"
mv "$tmp" "$BIN"
chmod +x "$BIN"

# 5. Sur macOS, retirer la mise en quarantaine (le programme n'est pas signe,
#    sans ca le systeme pourrait refuser de le lancer). L'erreur est sans
#    consequence si l'attribut n'existe pas.
if [ "$OS" = "darwin" ]; then
  xattr -d com.apple.quarantine "$BIN" 2>/dev/null || true
fi

echo "Programme installe dans $BIN"

# 6. Verifier que ~/.local/bin est dans le PATH.
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    echo ""
    echo "Note : $BIN_DIR n'est pas dans ton PATH."
    echo "Pour pouvoir taper 'mcp-slack' directement, ajoute cette ligne a la fin de ton fichier ~/.zshrc (ou ~/.bashrc) :"
    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "puis ouvre un nouveau terminal. Ce n'est pas obligatoire pour la suite."
    echo ""
    ;;
esac

# 7. Enregistrer mcp-slack dans Claude Code.
if command -v claude >/dev/null 2>&1; then
  # Retirer une eventuelle entree existante pour pouvoir relancer sans doublon.
  claude mcp remove mcp-slack >/dev/null 2>&1 || true
  if claude mcp add mcp-slack --scope user -- "$BIN"; then
    echo "mcp-slack a ete ajoute a Claude Code."
  else
    echo "Attention : l'ajout automatique a Claude Code a echoue."
    print_manual_config
  fi
else
  echo ""
  echo "La commande 'claude' n'a pas ete trouvee sur cette machine."
  print_manual_config
fi

# 8. Connexion a Slack (la fenetre s'ouvre tout de suite).
echo ""
if [ -r /dev/tty ]; then
  printf "Sous-domaine de ton espace Slack (par exemple 'etf1' pour etf1.slack.com) : " > /dev/tty
  read workspace < /dev/tty || workspace=""
  if [ -n "${workspace:-}" ]; then
    echo "Ouverture d'une fenetre pour te connecter a $workspace.slack.com..."
    if ! "$BIN" login "$workspace"; then
      echo "La connexion n'a pas abouti. Tu peux la relancer quand tu veux avec :"
      echo "    mcp-slack login $workspace"
    fi
  else
    echo "Aucun sous-domaine saisi. Lance la connexion quand tu es pret avec :"
    echo "    mcp-slack login <workspace>"
  fi
else
  echo "Installation non interactive : la connexion a Slack n'a pas ete lancee."
  echo "Lance-la toi-meme avec :"
  echo "    mcp-slack login <workspace>"
  echo "Par exemple : mcp-slack login etf1"
fi

echo ""
echo "Termine. Redemarre Claude (ou recharge la fenetre VS Code) pour qu'il voie mcp-slack."
