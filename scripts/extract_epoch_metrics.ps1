param(
    [Parameter(Mandatory = $true)]
    [string]$Log,
    [Parameter(Mandatory = $true)]
    [string]$Out,
    [string]$Language,
    [string]$EpochRegex = '(?i)\bepoch\b\s*[:=]?\s*(\d+)',
    [string]$SplitRegex = '(?i)\b(train|eval|validation|valid|dev|test)\b',
    [string]$KvRegex = '(?i)([a-z][a-z0-9_.-]*)\s*[:=]\s*([0-9]+(?:\.[0-9]+)?)%?',
    [string]$LanguageRegex = '(?i)(?:\blang(?:uage)?\b\s*[:=]\s*([a-z0-9_-]+)|\bfor\s+([a-z0-9_-]+))'
)

function Normalize-Split {
    param([string]$Split)
    switch ($Split.ToLower()) {
        "validation" { return "eval" }
        "valid" { return "eval" }
        "dev" { return "eval" }
        "test" { return "eval" }
        default { return $Split.ToLower() }
    }
}

function Split-From-Key {
    param([string]$Key)
    if ($Key -match '^(train)([._-]|$)') { return "train" }
    if ($Key -match '^(eval|validation|valid|dev|test)([._-]|$)') { return "eval" }
    return ""
}

function Trim-SplitPrefix {
    param([string]$Key, [string]$Split)
    return ($Key -replace "^$Split[._-]", "")
}

$epochPattern = [regex]::new($EpochRegex)
$splitPattern = [regex]::new($SplitRegex)
$kvPattern = [regex]::new($KvRegex)
$languagePattern = [regex]::new($LanguageRegex)

"epoch,metric,value,split,language" | Set-Content -Path $Out
$autoEpoch = 0

Get-Content -Path $Log | ForEach-Object {
    $line = $_
    $epoch = -1
    $epochMatch = $epochPattern.Match($line)
    if ($epochMatch.Success) {
        $epoch = [int]$epochMatch.Groups[1].Value
        if ($epoch -gt $autoEpoch) { $autoEpoch = $epoch }
    }

    $langValue = $Language
    if ([string]::IsNullOrWhiteSpace($langValue)) {
        $langMatch = $languagePattern.Match($line)
        if ($langMatch.Success) {
            if ($langMatch.Groups[1].Value) { $langValue = $langMatch.Groups[1].Value.ToLower() }
            elseif ($langMatch.Groups[2].Value) { $langValue = $langMatch.Groups[2].Value.ToLower() }
        }
    }

    $splitTokens = @()
    foreach ($match in $splitPattern.Matches($line)) {
        $splitTokens += [pscustomobject]@{
            Position = $match.Index
            Token    = (Normalize-Split $match.Groups[1].Value)
        }
    }

    $lineEpochAssigned = $epoch -ge 0
    foreach ($match in $kvPattern.Matches($line)) {
        $key = $match.Groups[1].Value.ToLower()
        $value = $match.Groups[2].Value
        $split = Split-From-Key -Key $key
        $metric = $key
        if ($split) {
            $metric = Trim-SplitPrefix -Key $metric -Split $split
        } else {
            $selected = $null
            foreach ($token in $splitTokens) {
                if ($token.Position -le $match.Index) { $selected = $token.Token } else { break }
            }
            if ($selected) { $split = $selected }
        }

        if (-not $split) { continue }
        $split = Normalize-Split $split
        $metric = $metric.Trim("._-")
        if ([string]::IsNullOrWhiteSpace($metric) -or $metric -eq $split) { $metric = "value" }

        if (-not $lineEpochAssigned) {
            $autoEpoch++
            $epoch = $autoEpoch
            $lineEpochAssigned = $true
        }

        "$epoch,$metric,$value,$split,$langValue" | Add-Content -Path $Out
    }
}
