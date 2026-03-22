{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    inputs:
    inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs { inherit system; };
        version = "0.0.0";

        toolchain = import ./nix/go.nix { inherit pkgs; };
        go = toolchain.go;

        aegis = import ./nix/aegis.nix { inherit pkgs go version; };
        docker = import ./nix/docker.nix { inherit pkgs aegis; };
        checks = import ./nix/checks.nix {
          inherit pkgs go;
          src = ./.;
        };

      in
      {
        devShells = {
          default = toolchain.devShell;
        };

        packages = {
          default = aegis;
          inherit aegis docker;
        };

        apps = {
          default = {
            type = "app";
            program = "${aegis}/bin/aegis";
          };
        };

        checks = checks;

        formatter = pkgs.nixfmt;
      }
    );
}
