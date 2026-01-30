param(
    [Parameter(Mandatory = $true)]
    [string]$Lang,
    [Parameter(Mandatory = $true)]
    [string]$Log,
    [switch]$Reverse,
    [switch]$Overwrite,
    [string[]]$TrainerArgs
)

$trainerPath = Join-Path $PSScriptRoot "../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel"
if (-not (Test-Path $trainerPath)) {
    Write-Error "External trainer not found at $trainerPath"
    Write-Error "This repo relies on an external trainer. Provide it or update the path."
    exit 1
}

$analysisDir = Join-Path $PSScriptRoot "../cmd/analysis2"
$scriptName = if ($Reverse) { "train_language_reverse.sh" } else { "train_language.sh" }
$scriptPath = Join-Path $analysisDir $scriptName

$extraArgs = @()
if ($Overwrite) {
    $extraArgs += "--overwrite"
}
if ($TrainerArgs) {
    $extraArgs += $TrainerArgs
}

$header = @(
    "[capture-training-log] language=$Lang reverse=$Reverse",
    "[capture-training-log] start: $(Get-Date -Format o)"
)
$header | Set-Content -Path $Log

& bash $scriptPath $Lang @extraArgs 2>&1 | Tee-Object -FilePath $Log -Append

"[capture-training-log] end: $(Get-Date -Format o)" | Add-Content -Path $Log
