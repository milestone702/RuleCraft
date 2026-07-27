param([Parameter(ValueFromPipeline=$true)]$inputJson)
$data = $inputJson | ConvertFrom-Json
$value = $data.value
$result = $value.ToLower()

Write-Output "{ `"value`": `"$result`" }"
