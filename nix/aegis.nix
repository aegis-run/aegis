{
  pkgs,
  go ? pkgs.go_1_26,
  version ? "0.0.0",
}:
pkgs.buildGoModule.override { inherit go; } {
  pname = "aegis";
  inherit version;
  src = pkgs.lib.cleanSource ../.;

  vendorHash = null;

  CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  doCheck = false;

  meta = {
    description = "Google Zanzibar-inspired ACL evaluation service";
    homepage = "https://github.com/aegis-run/aegis";
    license = pkgs.lib.licenses.mit;
    mainProgram = "aegis";
  };
}
