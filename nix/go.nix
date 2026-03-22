{
  pkgs,
  go ? pkgs.go_1_26,
}:
{
  inherit go;

  devShell = pkgs.mkShell {
    packages = with pkgs; [
      go
      gopls
      gotools
      delve
      golangci-lint
      govulncheck
      gosec

      just
      hl-log-viewer
    ];

    shellHook = ''
      set -e
      echo "Go:            $(go version | awk '{print $3, $4}')"
      echo "gopls:         $(gopls version | head -1)"
      echo "golangci-lint: $(golangci-lint version --short)"
      echo "govulncheck:   $(govulncheck -version)"
      echo "gosec:         $(gosec --version 2>&1)"
      echo "just:          $(just --version)"
    '';
  };
}
