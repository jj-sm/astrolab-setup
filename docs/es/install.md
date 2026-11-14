# Instalación

## Requisitos

- Una cuenta en un servidor Linux con acceso a internet (el servidor del
  laboratorio, o cualquier otra máquina a la que entres por SSH)
- `git`, `curl` y `wget`, que el servidor ya tiene
- Cerca de 1 GB libre en tu carpeta personal para Python, el entorno virtual y
  TinyTeX

No necesitas `sudo`: todo queda en tu carpeta personal.

## 1. Descarga el repositorio

Clónalo en tu carpeta personal. El script de instalación está pensado para
ejecutarse desde ahí, porque crea el entorno virtual en la carpeta actual
(`~/.ipynb`), que es donde el gestor busca Jupyter.

```bash
cd ~
git clone https://github.com/jj-sm/astrolab-setup.git
```

## 2. Elige un token

Jupyter y CARTA piden un token antes de dejar entrar a alguien. Define uno por
defecto en un archivo `.env`:

```bash
cp astrolab-setup/.env.example astrolab-setup/.env
nano astrolab-setup/.env        # define TOKEN=<algo aleatorio>
```

Usa solo letras, dígitos y `. _ ~ -`, porque el token va dentro de URLs. Para
generar uno aleatorio: `python3 -c "import secrets; print(secrets.token_urlsafe(24))"`.

La instalación copia este archivo a `~/.astroenvs.env` con permisos `600`, así
que solo tú puedes leerlo. git ignora `.env`; nunca lo subas al repositorio.

Puedes saltarte este paso: en ese caso el gestor te pedirá escribir un token
cada vez que inicies un servidor.

## 3. Ejecuta la instalación

```bash
bash astrolab-setup/ipynb_setup.bash
```

Hace estos pasos y se detiene en el primer error:

1. Instala `uv`
2. Instala Python 3.11 y 3.13 con `uv`
3. Muestra las versiones instaladas
4. Crea el entorno virtual `~/.ipynb` con Python 3.13
5. Instala `ipykernel` e `ipynb-jjsm-tools[full]` en él y registra el kernel
   *Python (ipynb 3.13)*
6. Lista los paquetes instalados
7. Instala TinyTeX, si no está ya
8. Instala los paquetes de LaTeX que necesita Matplotlib, si encuentra `tlmgr`
9. Descarga la última versión de `astroenvs` en `~/.local/bin/astroenvs`, copia
   `.env` a `~/.astroenvs.env` y `folder.bash` a `~/.folder.sh`, y agrega los
   alias a `~/.bashrc`, `~/.bash_profile` y `~/.profile`

Si en el paso 9 no puede llegar a GitHub, instala en su lugar la versión
anterior del gestor, escrita en bash (`astroenvs.sh`), así que igual tendrás un
`astroenvs` que funciona.

Cuando termine, carga los alias nuevos:

```bash
source ~/.bashrc
```

## 4. Comprueba que funciona

```bash
astroenvs --version     # astroenvs v1.0.0
astroenvs               # abre el menú
```

## Actualizar el gestor

No hace falta repetir toda la instalación. Descarga la última versión sobre el
binario anterior:

```bash
curl -fsSL https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-linux-amd64 \
  -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs
```

En un servidor ARM, cambia `linux-amd64` por `linux-arm64`.

Volver a ejecutar el script de instalación también es seguro: reinstala el
gestor y reescribe los alias sin duplicarlos.

## Carpetas de proyecto (`tfold`)

`tfold` crea la estructura habitual de un proyecto nuevo en la carpeta actual:

```bash
mkdir ~/mi-proyecto && cd ~/mi-proyecto
tfold          # crea out/ src/ data/ notebooks/, cada una con un .gitkeep
```
