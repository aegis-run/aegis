{
  pkgs,
  go ? pkgs.go_1_26,
  src,
}:
let
  cleanSrc = pkgs.lib.cleanSource src;
in
{
  tests = pkgs.buildGoModule.override { inherit go; } {
    pname = "aegis-tests";
    version = "0.0.0";
    src = cleanSrc;
    vendorHash = null;
    doCheck = true;
  };

  lint =
    pkgs.runCommand "aegis-lint"
      {
        buildInputs = [
          pkgs.golangci-lint
          go
        ];
      }
      ''
        cd ${cleanSrc}
        golangci-lint run ./...
        touch $out
      '';

  fmt =
    pkgs.runCommand "aegis-fmt"
      {
        buildInputs = [ go ];
      }
      ''
        diff <(gofmt -l ${cleanSrc}) <(echo -n "")
        touch $out
      '';

  vuln =
    pkgs.runCommand "aegis-vuln"
      {
        buildInputs = [
          pkgs.govulncheck
          go
        ];
      }
      ''
        cd ${cleanSrc}
        govulncheck ./...
        touch $out
      '';

  sec =
    pkgs.runCommand "aegis-sec"
      {
        buildInputs = [
          pkgs.gosec
          go
        ];
      }
      ''
        cd ${cleanSrc}
        gosec ./...
        touch $out
      '';
}
