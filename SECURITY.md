# Sécurité

## Ce que fait l'outil, et ce qu'il ne fait pas

`mcp-slack` donne à Claude un accès en lecture à ton propre Slack. Pour cela, il
réutilise la session de ton navigateur : un token (`xoxc`) et un cookie (`d`).

- Tout tourne en local, sur ta machine. Rien n'est envoyé à un serveur tiers.
- La session est rangée sur ton disque, dans le dossier de configuration de ton
  utilisateur (`<config utilisateur>/mcp-slack/auth.json`), en lecture par ton
  compte seulement.
- Chacun se connecte à son propre espace et ne voit que son propre Slack. Rien
  n'est partagé entre les personnes.
- Les outils d'écriture (poster un message, réagir) sont désactivés par défaut.

## Protéger ta session

Le token et le cookie équivalent à ta connexion Slack : traite le fichier
`auth.json` comme un mot de passe. Ne le copie pas, ne le colle pas dans un
message ou un ticket, ne le mets pas dans un dépôt. Si tu penses qu'il a fuité,
déconnecte ta session Slack (ce qui invalide le token et le cookie), puis relance
`mcp-slack login <workspace>` pour en obtenir une nouvelle.

## Signaler une faille

Si tu découvres une vulnérabilité, ne la publie pas dans une issue publique.
Ouvre un avis de sécurité privé sur le dépôt GitHub
(onglet « Security » > « Report a vulnerability »), ou contacte directement le
propriétaire du dépôt. Merci de laisser un délai raisonnable pour corriger avant
toute divulgation.
