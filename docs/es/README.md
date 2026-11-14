# astrolab-setup

`astrolab-setup` prepara tu cuenta en el servidor del laboratorio para trabajar
en astronomía y te da `astroenvs`, un gestor de terminal que inicia y detiene
Jupyter Lab, CARTA, DS9 y CASA por ti.

Todo se instala en tu carpeta personal, sin `sudo`.

Versión actual: 1.0.1 (ver el [changelog](changelog.md)).

| Página | Contenido |
|---|---|
| [Instalación](install.md) | el script de instalación, el token, cómo actualizar |
| [Uso del gestor](manager.md) | cada opción del menú, teclas, URLs, apps X11 |
| [Configuración](configuration.md) | opciones de `~/.astroenvs.env`, dónde quedan los logs |
| [Solución de problemas](troubleshooting.md) | errores comunes y cómo resolverlos |
| [Desarrollo](development.md) | estructura del código, compilación, versiones, esta documentación |
| [Changelog](changelog.md) | cambios por versión |

## Qué te deja instalado

- **Python** 3.11 y 3.13, instalados con [uv](https://docs.astral.sh/uv/)
- **`~/.ipynb`**, un entorno virtual de Python 3.13 registrado como el kernel de
  Jupyter *Python (ipynb 3.13)*, con
  [ipynb-jjsm-tools](https://github.com/jj-sm/ipynb-jjsm-tools) instalado
- **TinyTeX**, una distribución pequeña de LaTeX para el `usetex` de Matplotlib
  y la exportación a PDF
- **`astroenvs`**, el gestor, con los alias `astroenvs` e `ipynb`
- **`tfold`**, que crea las carpetas `out/`, `src/`, `data/` y `notebooks/` de
  un proyecto nuevo

## Uso rápido

En el servidor, desde tu carpeta personal:

```bash
git clone https://github.com/jj-sm/astrolab-setup.git
cp astrolab-setup/.env.example astrolab-setup/.env   # y luego define TOKEN=<algo aleatorio>
bash astrolab-setup/ipynb_setup.bash
source ~/.bashrc
astroenvs
```

Elige **Jupyter · Start server**, presiona Enter dos veces (tu carpeta personal
y el token por defecto) y abre la URL que aparece. Necesitas la VPN de la
universidad activa para llegar a ella.

```text
┃ ASTROENVS MANAGER
┃ Server 10.254.0.10 · v1.0.1
┃ > Jupyter  · Start server
┃   Jupyter  · Active instances
┃   Jupyter  · Logs
┃   CARTA    · Start server
┃   CARTA    · Active instances
┃   CARTA    · Logs
┃   DS9      · Open window (X11)
┃   CASA     · Start in this terminal
┃   Manage   · Stop an instance
┃   Manage   · Stop all instances
┃   Exit
↑ up • ↓ down • / filter • enter submit
```

El gestor está en inglés; esta documentación usa los nombres de las opciones tal
como aparecen en pantalla.
