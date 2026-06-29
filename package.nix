{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule {
  pname = "yarnfetch";
  version = "0-unstable-2026-06-29";

  src = fetchFromGitHub {
    owner = "yaaaarn";
    repo = "yarnfetch";
    rev = "57ea3bd22e103a1b3f74bf16683dbfb3f7cf6ead";
    hash = "sha256-NaksvcnnogbW9DMynrM3SshwLZFMC1DHYqJ0xG1Lf38=";
  };

  vendorHash = lib.fakeHash; # Nix will calculate the Go dependency hash

  meta = with lib; {
    description = "A simple info utility";
    homepage = "https://github.com/yaaaarn/yarnfetch";
    license = licenses.mit;
    mainProgram = "yarnfetch";
  };
}
