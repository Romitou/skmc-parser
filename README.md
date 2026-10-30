# skmc-parser

`skmc-parser` est un service en Go conçu pour parser et valider du code Skript sans avoir besoin d'exécuter un serveur Minecraft complet. Il fait office de passerelle HTTP et de répartiteur vers un ensemble de conteneurs Docker exécutant le moteur d'analyse.

## Pourquoi ce projet ?

Tester la syntaxe d'un script Skript nécessite habituellement de démarrer un serveur Spigot ou Paper avec le plugin Skript installé, de recharger le script via la console et de vérifier les erreurs dans les logs. Ce processus est lourd et peu adapté à une intégration continue ou à des outils en ligne (bots Discord, éditeurs web).

Ce service fournit une API HTTP simple : il reçoit le script en texte brut, choisit un conteneur d'analyse disponible via l'API Docker, lui transmet le code et retourne le résultat (erreurs et avertissements).

## Fonctionnement

- **Gestion des conteneurs** : le service se connecte au socket Docker local (`/var/run/docker.sock`) pour détecter les conteneurs actifs disposant du label `skriptmc.parser.engine-version` et reliés au réseau Docker `skmc-parser`.
- **Répartition des requêtes** : les requêtes d'analyse reçues sur `/parse` sont envoyées de manière aléatoire sur l'un des conteneurs sains détectés.
- **Limitation de débit (rate limiting)** : un middleware s'appuie sur une instance Redis pour limiter le nombre de requêtes par adresse IP et protéger le service contre les abus.
- **Tâches périodiques** : un scheduler interne met à jour la liste des conteneurs valides à intervalle régulier.

## API

### `GET /parsers`

Retourne la liste des conteneurs de parsing détectés et prêts à traiter des requêtes.

Exemple de réponse :
```json
{
  "lastUpdated": "2026-10-30T14:00:00Z",
  "parsers": [
    {
      "name": "/skmc-engine-1",
      "version": "1.0.0",
      "state": "running",
      "status": "Up 2 hours"
    }
  ]
}
```

### `POST /parse`

Envoie un script à analyser.

- **Header** : `Content-Type: text/plain`
- **Body** : Le code Skript brut.

Exemple de requête :
```bash
curl -X POST http://localhost:8000/parse \
  -H "Content-Type: text/plain" \
  --data-binary @mon_script.sk
```

Le service ajoute le header `X-Parser` indiquant quel conteneur a traité la demande et renvoie la réponse JSON fournie par le moteur.

## Configuration

La configuration par défaut se trouve dans `config.yaml`. Pour le développement local, il est possible de créer un fichier `config.local.yaml` (ignoré par Git) ou de surcharger les valeurs via des variables d'environnement préfixées par `SKMC_`.

| Clé YAML | Variable d'environnement | Défaut | Description |
| :--- | :--- | :--- | :--- |
| `http.listen` | `SKMC_HTTP_LISTEN` | `0.0.0.0` | Adresse d'écoute HTTP |
| `http.port` | `SKMC_HTTP_PORT` | `8000` | Port d'écoute du serveur |
| `rateLimiter.limit` | `SKMC_RATELIMITER_LIMIT` | `1000-H` | Quota autorisé (ex. 1000 requêtes par heure) |
| `rateLimiter.prefix` | `SKMC_RATELIMITER_PREFIX` | `limiter_skmcparser` | Préfixe des clés de quota dans Redis |
| `redis.address` | `SKMC_REDIS_ADDRESS` | `127.0.0.1:6379` | Adresse du serveur Redis |
| `redis.username` | `SKMC_REDIS_USERNAME` | `""` | Utilisateur Redis (si requis) |
| `redis.password` | `SKMC_REDIS_PASSWORD` | `""` | Mot de passe Redis |
| `redis.db` | `SKMC_REDIS_DB` | `0` | Base de données Redis |

## Lancer le projet

### Avec Docker

```bash
docker run -d \
  --name skmc-parser \
  --restart unless-stopped \
  --publish 8000:8000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --env SKMC_REDIS_ADDRESS="redis:6379" \
  --env SKMC_REDIS_PASSWORD="secretpassword" \
  ghcr.io/romitou/skmc-parser:latest
```

### En local

Prérequis : Go 1.25+ installé et une instance Redis accessible.

```bash
# Télécharger les dépendances
go mod download

# Lancer le serveur
go run main.go
```

## Licence

Projet sous licence [MIT](LICENSE).
