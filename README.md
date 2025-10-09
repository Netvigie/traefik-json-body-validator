# Traefik JSON Body Validator Plugin

Plugin middleware pour Traefik qui valide le corps des requêtes JSON avec des expressions régulières et des règles de validation.

## Fonctionnalités

- ✅ Validation par expressions régulières
- ✅ Vérification des champs requis
- ✅ Validation de longueur (min/max)
- ✅ Détection des valeurs vides
- ✅ Messages d'erreur personnalisables
- ✅ Support de plusieurs règles de validation

## Installation

### Configurer Traefik

#### Configuration statique (traefik.yml)

```yaml
experimental:
  plugins:
    json-body-validator:
      moduleName: github.com/username/traefik-json-body-validator
      version: v1.0.0
```

#### Configuration dynamique

```yaml
http:
  routers:
    my-router:
      rule: Path(`/api`)
      service: my-service
      entryPoints:
        - http
      middlewares:
        - json-validator

  services:
    my-service:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    json-validator:
      plugin:
        json-body-validator:
          rules:
            - field: userId
              pattern: "^.+$"
              required: true
          response:
            status: 400
            code: "INVALID_REQUEST"
            message: "userId is required and must have a non-empty value"
```

## Configuration

### Options de validation (rules)

Chaque règle supporte les options suivantes :

| Option | Type | Requis | Description |
|--------|------|--------|-------------|
| `field` | string | **Oui** | Nom du champ JSON à valider |
| `pattern` | string | Non | Expression régulière (regex) pour valider la valeur |
| `required` | boolean | Non | Si `true`, le champ doit être présent dans le JSON (défaut: `false`) |
| `minLength` | int | Non | Longueur minimale de la valeur |
| `maxLength` | int | Non | Longueur maximale de la valeur |

### Options de réponse (response)

Configuration de la réponse d'erreur :

| Option | Type | Description | Valeur par défaut |
|--------|------|-------------|-------------------|
| `status` | int | Code HTTP de la réponse d'erreur | `400` |
| `code` | string | Code d'erreur personnalisé | `"INVALID_REQUEST"` |
| `message` | string | Message d'erreur par défaut | `"Invalid request body"` |

## Exemples d'utilisation

### Exemple 1 : Valider un ID non vide

Vérifie que le champ `userId` existe et contient au moins un caractère.

```yaml
middlewares:
  json-validator:
    plugin:
      json-body-validator:
        rules:
          - field: userId
            pattern: "^.+$"
            required: true
        response:
          status: 400
          code: "MISSING_ID"
          message: "userId is required and cannot be empty"
```

**Requêtes valides :**
```json
{"userId": "abc123"}
{"userId": "user-456"}
```

**Requêtes invalides :**
```json
{}                       // Champ manquant
{"userId": ""}          // Valeur vide
```

### Exemple 2 : Valider plusieurs champs

```yaml
middlewares:
  json-validator:
    plugin:
      json-body-validator:
        rules:
          - field: email
            pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
            required: true
          - field: username
            pattern: "^[a-zA-Z0-9_]{3,20}$"
            required: true
            minLength: 3
            maxLength: 20
          - field: age
            pattern: "^[0-9]+$"
            required: false
        response:
          status: 422
          code: "VALIDATION_ERROR"
          message: "Request validation failed"
```

### Exemple 3 : Valider un UUID

```yaml
middlewares:
  json-validator:
    plugin:
      json-body-validator:
        rules:
          - field: userId
            pattern: "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
            required: true
          - field: requestId
            pattern: "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
            required: true
```

**Requête valide :**
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "requestId": "123e4567-e89b-12d3-a456-426614174000"
}
```

### Exemple 4 : Valider un numéro de téléphone

```yaml
middlewares:
  json-validator:
    plugin:
      json-body-validator:
        rules:
          - field: phone
            pattern: "^\\+?[1-9]\\d{1,14}$"
            required: true
            minLength: 10
            maxLength: 15
```

### Exemple 5 : Valider un mot de passe fort

```yaml
middlewares:
  json-validator:
    plugin:
      json-body-validator:
        rules:
          - field: password
            pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{8,}$"
            required: true
            minLength: 8
        response:
          status: 400
          message: "Password must contain at least 8 characters, including uppercase, lowercase, number and special character"
```

## Tests

### Tester avec curl

**Requête valide :**
```bash
curl -X POST http://localhost/api \
  -H "Content-Type: application/json" \
  -d '{"userId": "abc123"}'
```

**Requête invalide (champ manquant) :**
```bash
curl -X POST http://localhost/api \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Réponse d'erreur attendue :**
```json
{
  "error": "Field 'userId' is required",
  "code": "INVALID_REQUEST"
}
```

**Requête invalide (valeur vide) :**
```bash
curl -X POST http://localhost/api \
  -H "Content-Type: application/json" \
  -d '{"userId": ""}'
```

**Réponse d'erreur attendue :**
```json
{
  "error": "Field 'userId' cannot be empty",
  "code": "INVALID_REQUEST"
}
```

## Patterns regex utiles

Voici quelques expressions régulières courantes :

| Type | Pattern | Description |
|------|---------|-------------|
| Non vide | `^.+$` | Au moins un caractère |
| Alphanumérique | `^[a-zA-Z0-9]+$` | Lettres et chiffres uniquement |
| Email | `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` | Format email |
| UUID | `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$` | Format UUID v4 |
| Téléphone | `^\+?[1-9]\d{1,14}$` | Format E.164 |
| URL | `^https?://[^\s]+$` | URL HTTP/HTTPS |
| IP v4 | `^(\d{1,3}\.){3}\d{1,3}$` | Adresse IP v4 |
| Date ISO | `^\d{4}-\d{2}-\d{2}$` | Format YYYY-MM-DD |

## Dépannage

### Le plugin ne se charge pas

Vérifiez que :
1. Le repository GitHub est public
2. Le tag de version existe (`git tag -l`)
3. Le `moduleName` dans la configuration Traefik correspond exactement au chemin GitHub
4. La section `experimental.plugins` est bien dans la configuration statique

### Les requêtes ne sont pas validées

Vérifiez que :
1. Le middleware est bien attaché au router
2. Les requêtes ont bien un `Content-Type: application/json`
3. Le corps de la requête est bien au format JSON valide

### Erreur "invalid regex pattern"

Assurez-vous que :
1. Les barres obliques inverses sont échappées (`\\` au lieu de `\`)
2. La syntaxe regex est compatible avec Go (RE2)

## Contribuer

Les contributions sont les bienvenues ! N'hésitez pas à :
- Ouvrir une issue pour signaler un bug
- Proposer une pull request pour ajouter des fonctionnalités
- Améliorer la documentation

## Licence

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Support

Pour toute question ou problème :
- Ouvrez une issue sur GitHub
- Consultez la documentation Traefik : https://doc.traefik.io/traefik/plugins/
