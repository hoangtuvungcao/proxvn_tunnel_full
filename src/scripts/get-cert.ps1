# ProxVN Certificate Fingerprint Tool
# Lấy SHA256 fingerprint của server certificate

param(
    [string]$Server = "103.77.246.206",
    [int]$Port = 8882
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ProxVN Certificate Fingerprint Tool" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

try {
    Write-Host "[*] Connecting to $Server`:$Port..." -ForegroundColor Yellow
    
    # Create TCP connection
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $tcpClient.Connect($Server, $Port)
    $tcpStream = $tcpClient.GetStream()
    
    # Create SSL stream
    $sslStream = New-Object System.Net.Security.SslStream(
        $tcpStream, 
        $false,
        { param($sender, $certificate, $chain, $errors) return $true }
    )
    
    # Authenticate (get certificate)
    $sslStream.AuthenticateAsClient($Server)
    
    # Get remote certificate
    $remoteCert = $sslStream.RemoteCertificate
    
    if ($null -eq $remoteCert) {
        Write-Host "[!] Failed to get certificate" -ForegroundColor Red
        exit 1
    }
    
    # Calculate SHA256 fingerprint
    $certBytes = $remoteCert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Cert)
    $sha256 = [System.Security.Cryptography.SHA256]::Create()
    $hashBytes = $sha256.ComputeHash($certBytes)
    $fingerprint = -join ($hashBytes | ForEach-Object { $_.ToString("x2") })
    
    # Display results
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Certificate Information:" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Subject:     $($remoteCert.Subject)" -ForegroundColor White
    Write-Host "Issuer:      $($remoteCert.Issuer)" -ForegroundColor White
    Write-Host "Valid From:  $($remoteCert.GetEffectiveDateString())" -ForegroundColor White
    Write-Host "Valid To:    $($remoteCert.GetExpirationDateString())" -ForegroundColor White
    Write-Host ""
    Write-Host "SHA256 Fingerprint:" -ForegroundColor Yellow
    Write-Host $fingerprint -ForegroundColor Cyan
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Usage:" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "proxvn.exe --server $Server`:$Port \`" -ForegroundColor White
    Write-Host "           --cert-pin $fingerprint \`" -ForegroundColor Cyan
    Write-Host "           --proto http 3000" -ForegroundColor White
    Write-Host ""
    
    # Cleanup
    $sslStream.Close()
    $tcpStream.Close()
    $tcpClient.Close()
    
} catch {
    Write-Host ""
    Write-Host "[!] Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Common issues:" -ForegroundColor Yellow
    Write-Host "  - Server is not running" -ForegroundColor Gray
    Write-Host "  - Firewall blocking connection" -ForegroundColor Gray
    Write-Host "  - Wrong server address or port" -ForegroundColor Gray
    Write-Host ""
    exit 1
}
