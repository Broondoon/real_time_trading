# Define the path to env file as parameter as well as the path to docker-compose.yaml
param(
    [string]$envFilePath = ".env",
    [string]$dockerComposeFilePath = "docker-compose.yaml"
) 

# Check if the .env file exists
if (Test-Path $envFilePath) {
    # Read the contents of the .env file
    $envVariables = Get-Content $envFilePath

    # Loop through each line in the .env file
    foreach ($line in $envVariables) {
        # Split the line into variable name and value
        $parts = $line -split "="

        [System.Environment]::SetEnvironmentVariable($parts[0], ($parts[1..($parts.Length - 1)] -join "="), [System.EnvironmentVariableTarget]::Process)
    }
} else {
    Write-Host "The .env file does not exist."
}

# Input string containing references to environment variables
$inputString = Get-Content $dockerComposeFilePath -Raw

# Regular expression pattern to match environment variable references
$pattern = '\${([a-zA-Z_][a-zA-Z0-9_]*)}'

# Replace environment variable references with their values
$outputString = [regex]::Replace($inputString, $pattern, {
    param($match)

    $envValue = [System.Environment]::GetEnvironmentVariable($match.Groups[1].Value)
    if ($envValue -ne $null) {
        return $envValue
    } else {
        return $match.Value
    }
})

# Output the modified string
Write-Output $outputString