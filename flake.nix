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


      in
      {
        devShells = {
          default = toolchain.devShell;
        };

      }
    );
}
