# mcp-slack

Un serveur MCP qui donne à Claude un accès **en lecture** à **ton propre** Slack.

Il réutilise la session de ton navigateur, déjà connectée à Slack. Il n'y a donc
aucune application à faire approuver par un administrateur, aucun bot à créer,
aucun token à copier à la main.

## Ce que ça fait, concrètement

- Claude peut lire tes canaux, tes fils de discussion et tes messages directs,
  chercher dans les messages, lister les canaux et les personnes.
- Il lit **ta** session Slack : un token et un cookie pris dans ton navigateur.
- Tout tourne **en local**, sur ta machine. Rien n'est envoyé ailleurs.
- **Chacun se connecte à son propre espace et ne voit que son propre Slack.**
  Rien n'est partagé avec qui que ce soit.
- Écrire dans Slack (poster, réagir) est désactivé par défaut.

Pour le détail sur la session et la confidentialité, voir [SECURITY.md](SECURITY.md).

## Installation (Claude Code, y compris l'extension VS Code)

Ouvre un terminal et colle cette ligne :

```sh
curl -fsSL https://raw.githubusercontent.com/florianzilliox/mcp-slack/main/install.sh | sh
```

Ce qui se passe ensuite :

1. Le script télécharge le bon programme pour ta machine et l'installe.
2. Il te demande le sous-domaine de ton espace Slack (par exemple `etf1` pour
   `etf1.slack.com`).
3. Une fenêtre de navigateur s'ouvre : connecte-toi à Slack normalement (SSO
   compris). La fenêtre se ferme toute seule une fois la connexion récupérée.
4. Redémarre Claude (ferme et rouvre la fenêtre VS Code si tu utilises
   l'extension). C'est prêt.

### Ou : demande à ton Claude de l'installer

Dans Claude Code (dans VS Code aussi), tu peux simplement écrire :

> Installe le serveur MCP Slack depuis https://github.com/florianzilliox/mcp-slack
> en suivant son README.

Claude lit ce fichier et lance l'installation à ta place. Il te demandera juste
le sous-domaine de ton espace Slack, puis une fenêtre s'ouvrira pour te connecter.

## Se reconnecter plus tard

Si ta session Slack expire, ou pour te connecter à un autre espace, relance
simplement la connexion :

```sh
mcp-slack login <workspace>
```

Par exemple : `mcp-slack login etf1`. Une fenêtre s'ouvre, tu te reconnectes, et
c'est reparti.

## Désinstaller

Retire l'entrée de Claude Code puis supprime le programme :

```sh
claude mcp remove mcp-slack
rm ~/.local/bin/mcp-slack
```
