{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self, ... }@inputs:
    inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs { inherit system; };

        version =
          if self ? rev then
            if self ? ref then
              let
                tag = builtins.replaceStrings [ "refs/tags/" ] [ "" ] self.ref;
              in
              if builtins.match "v[0-9].*" tag != null then tag else "v0.0.0-${builtins.substring 0 7 self.rev}"
            else
              "v0.0.0-${builtins.substring 0 7 self.rev}"
          else
            "v0.0.0-dirty";

        toolchain = import ./nix/go.nix { inherit pkgs; };
        go = toolchain.go;

        aegis = import ./nix/aegis.nix {
          inherit
            self
            pkgs
            go
            version
            ;
        };

      in
      {
        devShells.default = toolchain.devShell;

        packages = {
          default = aegis;
          inherit aegis;
        };

        apps.default = {
          type = "app";
          program = "${aegis}/bin/aegis";
          meta = aegis.meta;
        };

        formatter = pkgs.nixfmt;
      }
    );
}
