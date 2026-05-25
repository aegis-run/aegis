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

  vendorHash = "sha256-588/zfA5y2IZf7kSRqpUNkV9TkSQ4bPU+xReTORMIKw=";

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
