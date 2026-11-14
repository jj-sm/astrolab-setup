# Changelog

## Sin publicar

- Instalación con un solo comando: `curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash`
  descarga el repositorio en una carpeta temporal, pide un token, ejecuta la
  instalación y borra la carpeta. No necesita `git`.
- Volver a ejecutar la instalación reutiliza un `~/.ipynb` existente en vez de
  detenerse en la pregunta "already exists" de uv, y el entorno siempre se crea
  en tu carpeta personal, desde donde sea que ejecutes la instalación.
- La instalación busca `.env` y `folder.bash` junto al script, así que copia tu
  token a `~/.astroenvs.env` aunque la ejecutes desde otra carpeta. Avisa cuando
  ese archivo no tiene `TOKEN`.

## 1.0.1 (2026-11-14)

- Se eliminó el gestor en bash (`astroenvs.sh`). La instalación solo instala el
  binario en Go; si la descarga falla, termina todo lo demás y se detiene
  mostrando el comando para descargarlo a mano.

## 1.0.0 (2026-11-14)

El gestor ahora es un programa en Go, `astroenvs`, que el script de instalación
descarga como binario.

- Cada pregunta es un selector: flechas y Enter, Esc para volver
- Jupyter pregunta en qué carpeta abrirse (tu carpeta personal por defecto) y la
  usa como tope de su explorador de archivos
- Cada servidor de Jupyter y CARTA puede usar el token por defecto o uno que
  escribas
- Las rutas personalizadas se escriben con autocompletado con Tab
- DS9 lista los archivos FITS de la carpeta elegida
- Los servidores de Jupyter se registran como los de CARTA: puerto, token y log
  propios, con la URL a la vista
- Los logs se eligen de una lista; **Stop an instance** elige de una lista en
  vez de pedir un PID
- La configuración pasó del inicio del script a `~/.astroenvs.env`
- Se siguen reconociendo las instancias que inició el gestor en bash
- El script de instalación usa el gestor en bash si no puede llegar a GitHub
