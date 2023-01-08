#/bin/sh

set -x 
function test {
  echo "+ $@"
  "$@"
  local status=$?
  if [ $status -ne 0 ]; then
    exit $status
  fi
  return $status
}

if [ -d "${GOPATH}/src/github.com/elgatito/elementum" ];
then
  GIT_VERSION=`cd ${GOPATH}/src/github.com/elgatito/elementum; git describe --tags`
else
  GIT_VERSION=`git describe --tags`
fi

DEST_MAKE=linux-x64
DEST_PLATFORM=linux_x64
DEST_DIR=$HOME/.kodi
if [ ! -z "${WSL_USER}" ]; then
  DEST_DIR=/mnt/c/Users/${WSL_USER}/AppData/Roaming/Kodi
  DEST_PLATFORM=windows_x64
  DEST_MAKE=windows-x64
fi

LOCAL_ENV=$GOPATH/src/github.com/ElementumOrg/libtorrent-go/local-env/
if [ -d "$CROSS_ROOT" ];
then
    LOCAL_ENV=$CROSS_ROOT
fi

# This will run with local go using libtorrent-go/local-env/ locally copied dependencies compilation.
export LOCAL_ENV=$LOCAL_ENV
export PATH=$PATH:$LOCAL_ENV/bin/
export PKG_CONFIG_PATH=$LOCAL_ENV/lib/pkgconfig
export SWIG_LIB=$LOCAL_ENV/share/swig/4.1.0/

if [ "$1" == "local" ]
then
  set -e
  test go build -ldflags="-w -X github.com/elgatito/elementum/util.Version=${GIT_VERSION}" -o /var/tmp/elementum .
  test chmod +x /var/tmp/elementum*
  test cp -rf /var/tmp/elementum* $DEST_DIR/addons/plugin.video.elementum/resources/bin/$DEST_PLATFORM/
  test cp -rf /var/tmp/elementum* $DEST_DIR/userdata/addon_data/plugin.video.elementum/bin/$DEST_PLATFORM/
elif [ "$1" == "library" ]
then
  set -e
  test go build -ldflags="-w -X github.com/elgatito/elementum/util.Version=${GIT_VERSION}" -tags shared -buildmode=c-shared -o /var/tmp/elementum.so .
  test chmod +x /var/tmp/elementum*
  test cp -rf /var/tmp/elementum* $DEST_DIR/addons/plugin.video.elementum/resources/bin/$DEST_PLATFORM/
  test cp -rf /var/tmp/elementum* $DEST_DIR/userdata/addon_data/plugin.video.elementum/bin/$DEST_PLATFORM/
elif [ "$1" == "sanitize" ]
then
  # This will run with local go
  set -e
  CGO_ENABLED=1 CGO_LDFLAGS='-fsanitize=leak -fsanitize=address' CGO_CFLAGS='-fsanitize=leak -fsanitize=address' test go build -ldflags="-w -X github.com/elgatito/elementum/util.Version=${GIT_VERSION}" -o /var/tmp/elementum github.com/elgatito/elementum
  test chmod +x /var/tmp/elementum*
  test cp -rf /var/tmp/elementum* $DEST_DIR/addons/plugin.video.elementum/resources/bin/$DEST_PLATFORM/
  test cp -rf /var/tmp/elementum* $DEST_DIR/userdata/addon_data/plugin.video.elementum/bin/$DEST_PLATFORM/
elif [ "$1" == "docker" ]
then
  # This will run with docker libtorrent:$DEST_MAKE image
  test make $DEST_MAKE
  test cp -rf build/$DEST_PLATFORM/elementum* $DEST_DIR/addons/plugin.video.elementum/resources/bin/$DEST_PLATFORM/
  test cp -rf build/$DEST_PLATFORM/elementum* $DEST_DIR/userdata/addon_data/plugin.video.elementum/bin/$DEST_PLATFORM/
elif [ "$1" == "docker-library" ]
then
  # This will run with docker libtorrent:$DEST_MAKE image
  test make ${DEST_MAKE}-shared
  test cp -rf build/${DEST_PLATFORM}/elementum.* $DEST_DIR/addons/plugin.video.elementum/resources/bin/$DEST_PLATFORM/
  test cp -rf build/${DEST_PLATFORM}/elementum.* $DEST_DIR/userdata/addon_data/plugin.video.elementum/bin/$DEST_PLATFORM/
fi
