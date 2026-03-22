{ pkgs, aegis }:
pkgs.dockerTools.buildLayeredImage {
  name = "aegis";
  tag = aegis.version;
  contents = [
    pkgs.cacert
    aegis
  ];
  config = {
    Entrypoint = [ "/bin/aegis" ];
    ExposedPorts = {
      "8080/tcp" = { };
    };
    User = "65534:65534";
    Env = [
      "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
    ];
  };
}
