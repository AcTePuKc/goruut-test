Param(
  [Parameter(Mandatory = $true)][string]$Lang,
  [string]$Log = "./graph_run.log",
  [switch]$Gist,
  [int]$GistK = 0,
  [double]$GistLambda = 1.0
)

# Tiny runner stub for the training/eval workflow.
# NOTE: This repo calls external trainer binaries located at ../../../classifier.
# If those binaries are missing, this script will fail.

$ErrorActionPreference = "Stop"

"[graph-runner] language=$Lang" | Tee-Object -FilePath $Log
"[graph-runner] gist_enable=$Gist gist_k=$GistK gist_lambda=$GistLambda" | Tee-Object -FilePath $Log -Append
"[graph-runner] start: $(Get-Date -Format o)" | Tee-Object -FilePath $Log -Append

# TODO: Provide a Windows-native runner or use WSL/Git-Bash to execute ./cmd/analysis2/*.sh scripts.
# Example using WSL (requires bash + external trainer):
# wsl bash -lc "cd /workspace/goruut-test/cmd/analysis2 && ./study_language.sh $Lang"

"[graph-runner] TODO: run cmd/analysis2 scripts via WSL or provide a Windows-native equivalent." |
  Tee-Object -FilePath $Log -Append

"[graph-runner] end: $(Get-Date -Format o)" | Tee-Object -FilePath $Log -Append
