# Mongo Promote Tool

## Project docs

- [Working rules](AGENTS.md): mandatory coding, safety, and testing rules to review before changing code.
- [Architecture](docs/ARCHITECTURE.md): proposed Go package boundaries and first milestone.
- [Testing environment](docs/TESTING.md): local MongoDB setup for disposable integration tests.

## Idea

Esta herramienta compara dos bases de datos MongoDB y genera un plan controlado para llevar datos configurables desde una base origen hacia una base destino.

El caso principal es:

```txt
staging -> production
```

Pero la herramienta debe ser generica para poder usarse con cualquier par de bases MongoDB:

```txt
dev -> staging
staging -> production
backup -> production
client-a -> client-b
```

La herramienta no debe asumir que toda la base origen es la verdad. MongoDB no distingue por si solo entre datos de configuracion, datos de prueba y datos reales de negocio. Por eso el comportamiento debe estar controlado por un archivo de reglas.

## Lenguaje elegido

La herramienta se va a desarrollar en **Go**.

Motivos:

- Permite distribuir un binario unico por sistema operativo.
- No obliga a instalar Node.js, Python ni dependencias del proyecto donde se use.
- Es una buena opcion para CLIs operativas.
- Tiene driver oficial de MongoDB mantenido por MongoDB.
- Maneja concurrencia de forma simple si mas adelante se quieren comparar collections en paralelo.
- Permite tipar estructuras internas como config, plan, operaciones y auditoria sin depender del stack de la aplicacion que consume la herramienta.
- Es adecuado para una herramienta reutilizable entre distintos proyectos y equipos.

La herramienta debe ser independiente del backend actual. Aunque este repositorio use Node.js, esta CLI no debe depender de ese runtime.

## Que problema resuelve

Resuelve la promocion controlada de datos entre dos MongoDB sin pisar datos productivos.

Ejemplos de datos que normalmente si se pueden promover:

```txt
settings
roles
permissions
vehicleTypes
documentTemplates
formFields
languages
paymentGateways
billingPlans
subscriptionPlans
contentPages
categories base
```

Ejemplos de datos que normalmente no se deben promover:

```txt
users
orders
payments
walletTransactions
cards
trips
logs
notifications
reviews
chats
sessions
tokens
```

## Que no debe intentar resolver

La herramienta no debe ser un boton de:

```txt
hacer production igual a staging
```

Eso es peligroso porque staging suele tener datos de prueba, datos incompletos, IDs distintos y cambios experimentales.

La herramienta debe trabajar con:

- collections permitidas explicitamente;
- claves estables por collection;
- campos ignorados;
- reglas de arrays;
- validaciones antes de aplicar;
- modo dry-run por defecto;
- confirmacion explicita para cambios reales.

## Conceptos clave

### Source

Base de datos origen. Normalmente staging.

### Target

Base de datos destino. Normalmente production.

### Plan

Archivo generado por la herramienta que contiene las operaciones necesarias para actualizar el target.

Ejemplos:

```txt
insert document
update document
create index
drop index, solo si esta explicitamente permitido
delete document, solo si esta explicitamente permitido
```

### Stable key

Clave logica usada para identificar el mismo documento en ambas bases.

No se recomienda usar `_id` por defecto, porque el mismo concepto puede tener IDs distintos entre ambientes.

Ejemplos:

```txt
settings.key
roles.slug
languages.code
vehicleTypes.code
documentTemplates.slug
paymentGateways.provider
```

Tambien puede ser una clave compuesta:

```txt
country + city + serviceType
```

## Archivo de configuracion

Ejemplo de `mongo-promote.yml`:

```yaml
source:
  uriEnv: STAGING_MONGO_URI

target:
  uriEnv: PRODUCTION_MONGO_URI

defaults:
  dryRun: true
  allowDeletes: false
  ignoreFields:
    - _id
    - __v
    - createdAt
    - updatedAt

collections:
  settings:
    key:
      - key
    mode: upsert
    ignoreFields:
      - lastModifiedBy

  roles:
    key:
      - slug
    mode: upsert

  vehicleTypes:
    key:
      - code
    mode: upsert
    arrays:
      permissions:
        strategy: sort
      zones:
        strategy: preserveOrder

  documentTemplates:
    key:
      - slug
    mode: upsert

indexes:
  settings:
    - keys:
        key: 1
      options:
        unique: true
        name: key_unique

  roles:
    - keys:
        slug: 1
      options:
        unique: true
        name: slug_unique
```

## Comandos esperados

### Validar configuracion

```bash
mongo-promote validate --config=mongo-promote.yml
```

Debe revisar:

- que existan las variables de entorno;
- que se pueda conectar a ambas bases;
- que las collections configuradas existan o puedan crearse;
- que cada collection tenga una clave estable;
- que no haya reglas ambiguas.

### Generar diff

```bash
mongo-promote diff --config=mongo-promote.yml
```

Debe mostrar diferencias sin tocar la base destino.

Ejemplo:

```txt
settings
  + insert key="dynamic_pricing_enabled"
  ~ update key="commission_percentage"
    value: 10 -> 12

roles
  + insert slug="dispatcher"

indexes
  + createIndex settings.key_unique
```

### Generar plan

```bash
mongo-promote plan --config=mongo-promote.yml --out=plans/2026-07-29-staging-to-prod.json
```

Debe crear un archivo JSON con operaciones concretas, por ejemplo:

```json
{
  "source": "staging",
  "target": "production",
  "createdAt": "2026-07-29T00:00:00.000Z",
  "operations": [
    {
      "type": "updateOne",
      "collection": "settings",
      "filter": {
        "key": "commission_percentage"
      },
      "update": {
        "$set": {
          "value": 12
        }
      },
      "options": {
        "upsert": true
      }
    }
  ]
}
```

### Aplicar plan

```bash
mongo-promote apply --plan=plans/2026-07-29-staging-to-prod.json --confirm=production
```

Debe aplicar solamente las operaciones presentes en el plan.

Por seguridad, `apply` no deberia recalcular el diff en ese momento. Primero se genera el plan, se revisa, y despues se aplica ese plan exacto.

### Exportar migracion

```bash
mongo-promote export-script --plan=plans/2026-07-29-staging-to-prod.json --format=mongodb-js
```

Debe generar un script revisable para guardar en Git cuando haga falta.

Como la herramienta estara hecha en Go, la ejecucion principal no depende de `migrate-mongo`. Aun asi, podria soportar exportar scripts compatibles con otros flujos si un proyecto especifico los necesita.

## Flujo recomendado

```txt
1. Se hacen cambios en staging.
2. Se ejecuta diff contra production.
3. Se revisan diferencias.
4. Se genera un plan.
5. Se hace backup o snapshot de production.
6. Se aplica el plan en production con confirmacion explicita.
7. Se guarda el plan aplicado como evidencia.
8. Opcionalmente se exporta como migracion versionada.
```

## Restricciones importantes

### Dry-run por defecto

Ningun comando debe modificar datos salvo que se pase una bandera explicita:

```bash
--apply
--confirm=production
```

### Allowlist obligatoria

La herramienta no debe comparar todas las collections por defecto.

Solo debe operar sobre collections definidas en el archivo de configuracion.

### Deletes desactivados por defecto

Si un documento existe en target pero no existe en source, la herramienta puede reportarlo, pero no debe borrarlo salvo que la config lo permita:

```yaml
allowDeletes: true
```

Incluso con deletes habilitados, deberia poder configurarse por collection.

### No usar `_id` por defecto

`_id` debe ignorarse por defecto.

Solo se debe usar `_id` como clave si la config lo declara explicitamente.

### Indexes unicos requieren validacion previa

Antes de crear un index `unique`, la herramienta debe validar duplicados en target.

Si hay duplicados, debe fallar antes de aplicar.

### Referencias entre collections

MongoDB no valida relaciones como una base SQL. Si un documento tiene referencias a otra collection, la herramienta debe poder:

- mantener el valor tal cual;
- mapear por clave estable;
- avisar si la referencia no existe en target;
- bloquear la operacion si la referencia es requerida.

### Arrays necesitan estrategia

No todos los arrays se comparan igual.

Estrategias posibles:

```txt
preserveOrder: el orden importa
sort: se ordena antes de comparar
set: se compara como conjunto sin duplicados
replace: se reemplaza el array completo
mergeByKey: se mergean objetos dentro del array por una clave interna
```

### Campos calculados o runtime

Campos como estos normalmente se ignoran:

```txt
createdAt
updatedAt
lastLogin
lastSeen
usageCount
runtimeStats
temporaryToken
cache
```

### Auditoria

Cada ejecucion real debe guardar un log:

```txt
fecha
usuario o machine user
source
target
plan aplicado
cantidad de inserts
cantidad de updates
cantidad de deletes
cantidad de indexes
errores
```

Ese log puede guardarse en archivo y tambien en una collection del target:

```txt
mongo_promote_runs
```

## Modos de comparacion

### Document-level diff

Compara documentos completos despues de normalizar campos ignorados.

Bueno para configuraciones chicas.

### Field-level diff

Genera `$set` y `$unset` por campo.

Bueno para evitar reemplazar documentos completos.

### Index diff

Compara indexes esperados contra indexes existentes.

Debe crear indexes faltantes, pero no borrar indexes extra salvo configuracion explicita.

## Operaciones soportadas

Inicialmente:

```txt
insertOne
updateOne with $set
updateOne with $unset
createIndex
```

Mas adelante:

```txt
deleteOne
dropIndex
rename field
transform field
map references
merge arrays by key
```

## Protecciones recomendadas

- bloquear ejecucion si source y target apuntan a la misma URI;
- pedir `--confirm=production` para aplicar en production;
- mostrar conteos antes y despues;
- soportar `--max-operations` para evitar cambios masivos accidentales;
- soportar `--collections` para limitar una corrida;
- fallar si una collection configurada no tiene clave estable;
- fallar si hay claves duplicadas en source o target;
- no imprimir secretos de conexion en logs;
- guardar plan antes de aplicar;
- aplicar operaciones de forma ordenada;
- soportar rollback solamente cuando sea realmente posible.

## Estructura sugerida del proyecto Go

```txt
mongo-promote/
  cmd/
    mongo-promote/
      main.go
  internal/
    cli/
      root.go
      validate.go
      diff.go
      plan.go
      apply.go
      export.go
    config/
      config.go
      load.go
      validate.go
    mongo/
      connect.go
      collections.go
      indexes.go
    diff/
      normalize.go
      documents.go
      collections.go
      indexes.go
    plan/
      operation.go
      build.go
      read.go
      write.go
    apply/
      apply.go
      validators.go
    audit/
      audit.go
    exporters/
      mongodb_js.go
  examples/
    mongo-promote.yml
  go.mod
  go.sum
  README.md
```

## Resultado final de la idea

La idea queda como una CLI generica para MongoDB que permite comparar dos bases y promover cambios controlados desde una hacia otra.

No reemplaza backups.
No reemplaza migraciones estructurales versionadas.
No debe tocar datos productivos sin reglas explicitas.

Su rol ideal es:

```txt
promover datos configurables y metadata entre ambientes MongoDB de forma segura, auditable y repetible.
```

Para cambios que dependen del codigo, conviene seguir usando migraciones versionadas.

Para cambios hechos desde un panel admin en staging, esta herramienta puede generar el diff, plan y apply hacia production.

La implementacion elegida sera Go para que la herramienta quede desacoplada del lenguaje de cada backend y pueda distribuirse como binario reutilizable.

