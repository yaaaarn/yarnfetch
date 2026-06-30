{
  lib,
  buildGoModule,
}:
buildGoModule {
  pname = "yarnfetch";
  version = "unstable";

  src = ./.;

  vendorHash = "sha256-WZTMj4x4BNAMS3PFEkvPHE1md6sLO1IWvK0yvG9drCM=";

  meta = with lib; {
    description = "a simple info utility";
    homepage = "https://github.com/yaaaarn/yarnfetch";
    license = licenses.mit;
    mainProgram = "yarnfetch";
  };
}
