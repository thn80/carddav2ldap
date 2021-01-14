with import <nixpkgs> {};

mkShell {
  name = "carddav2ldap";
  nativeBuildInputs = [
    go
    goimports
    stdenv
    xmlformat
    jq
    gitFull
#    glibc.static
    devd
    ldapvi
  ];

  CFLAGS="-I${pkgs.glibc.dev}/include";
  LDFLAGS="-L${pkgs.glibc}/lib";
  GOOS="linux";
#  GOARCH="arm";
#  GOARM="7";

  shellHook = ''
    export GOPATH="$(pwd)/go"
    mkdir -pv "$GOPATH"
    export GOBIN="$GOPATH/bin"
    mkdir -pv "$GOBIN"
    export GOCACHE="$GOPATH/cache"
    mkdir -pv "$GOCACHE"
    export PATH="$PATH:$GOROOT/bin:$GOBIN"
    mkdir -pv src

    echo -n "Enter your helsinki username(SOGo): "
    read -r HELSINKI_USER
    export HELSINKI_USER
    
    echo -n "Enter your helsinki password(SOGo): "
    read -r -s HELSINKI_PASSWORD
    export HELSINKI_PASSWORD
    cd "src"
  '';
}
