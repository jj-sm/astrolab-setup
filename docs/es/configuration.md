# Configuración

## El archivo de configuración

El gestor lee su configuración del primero de estos archivos que exista:

1. el archivo indicado en la variable de entorno `ASTROENVS_ENV`
2. `~/.astroenvs.env` (lo crea la instalación a partir de tu `.env`)
3. `.env` junto al ejecutable `astroenvs`

Cada opción es una línea `CLAVE=valor`. Se ignoran las líneas vacías y las que
empiezan con `#`, y también las comillas alrededor del valor y un `export` al
inicio:

```bash
TOKEN=k3Jq9-wX2v_LpA
JUPYTER_ROOT=~/notebooks
DATA_DIRS=/data:~/UC-AST221-Rsync:~/Tareas
CARTA_BIN=/opt/carta/bin/carta
```

Una variable de entorno con el mismo nombre tiene prioridad sobre el archivo,
así que puedes probar un valor en una sola ejecución sin editar nada:

```bash
DATA_DIRS=/data:/scratch/$USER astroenvs
```

`TOKEN` es la excepción: solo se lee del archivo, porque una variable `TOKEN`
en tu terminal probablemente es para otra cosa.

El gestor lee el archivo al iniciar. Reinícialo después de editarlo.

## Opciones

| Opción | Por defecto | Qué hace |
|---|---|---|
| `TOKEN` | *(ninguno)* | token por defecto de Jupyter y CARTA; sin él escribes uno cada vez |
| `JUPYTER_ROOT` | `~` | carpeta que se ofrece primero (y por defecto) al iniciar Jupyter |
| `DATA_DIRS` | `/data:~/UC-AST221-Rsync` | carpetas de inicio que se ofrecen para todas las herramientas, separadas por `:` |
| `IPYNB_VENV` | `~/.ipynb` | entorno virtual que provee `jupyter` |
| `CARTA_BIN` | `carta` del `PATH` | ejecutable de CARTA |
| `DS9_BIN` | `ds9` del `PATH` | ejecutable de DS9 |
| `CASA_BIN` | `casa` del `PATH`, si no `/usr/local/casa/bin/casa` | ejecutable de CASA |
| `CARTA_TOP` | `/` | el explorador de archivos de CARTA no puede subir más arriba de esta carpeta |
| `JUPYTER_PORT_START` | `8888` | primer puerto que se prueba para Jupyter |
| `CARTA_PORT_START` | `3002` | primer puerto que se prueba para CARTA |
| `SERVER_IP` | primera dirección de red | dirección que aparece en las URLs |

Las rutas pueden empezar con `~`. Define `SERVER_IP` si las URLs muestran una
dirección equivocada, lo que puede pasar en máquinas con varias interfaces de
red.

## Tokens

Un token no puede estar vacío y solo puede usar letras, dígitos y `. _ ~ -`. El
gestor no inicia un servidor sin token, porque quedaría abierto para cualquiera
que pueda llegar a él.

El token por defecto es el del archivo de configuración. Al iniciar un servidor
puedes escribir otro; se usa solo en ese servidor y aparece en su URL.

## Archivos que escribe el gestor

Todo queda en `~/.astrolab/`, así que el gestor funciona igual desde cualquier
carpeta:

| Ruta | Contenido |
|---|---|
| `~/.astrolab/logs/jupyter-<puerto>.log` | salida de cada servidor de Jupyter |
| `~/.astrolab/logs/carta-<puerto>.log` | salida de cada servidor de CARTA |
| `~/.astrolab/logs/ds9-<pid>.log` | salida de cada ventana de DS9 |
| `~/.astrolab/run/<tipo>-<id>.json` | un archivo por instancia activa: PID, puerto, token, carpeta, log |

Los archivos de `run/` se borran cuando su instancia se detiene, y los que
quedan obsoletos se limpian solos. Los logs se conservan; borra los antiguos
cuando quieras. CASA escribe sus propios archivos `casa-*.log` en la carpeta
donde lo iniciaste.
