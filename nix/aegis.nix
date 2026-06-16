{
  self,
  pkgs,
  go ? pkgs.go_1_26,
  version ? "0.0.0",
}:
pkgs.buildGoModule.override { inherit go; } {
  pname = "aegis";
  inherit version;
  src = pkgs.lib.cleanSource ../.;

  vendorHash = "sha256-w69qsWbwpObFAJsFXDG8o7Ga1lO2ApOGGZ3nm8Aw0ZU=";

  ldflags = [
    "-s"
    "-w"
    "-X github.com/aegis-run/aegis/internal.Version=${version}"
    "-X github.com/aegis-run/aegis/internal.Identifier=${builtins.substring 0 7 (self.rev or "dirty")}"
  ];

  doCheck = false;

  meta = {
    description = "Open-source Centralized Authorization System";
    homepage = "https://github.com/aegis-run/aegis";
    license = pkgs.lib.licenses.mit;
    mainProgram = "aegis";
  };
}
