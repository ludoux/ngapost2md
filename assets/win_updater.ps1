$update_download_dir = Join-Path (Get-Location) "\update_temp\";
$ps1_filename = "win_updater.ps1"

function Override-Copy($source, $dest) {
    if (Test-Path ($source)) {
        Copy-Item $source -Destination $dest -force
    }
}

function Download-Archive ($filename, $link) {
    Write-Host "Downloading" $filename "..."
    Invoke-WebRequest -Uri $link -UserAgent [Microsoft.PowerShell.Commands.PSUserAgent]::FireFox -OutFile (Join-Path $update_download_dir $filename)
}

function Get-Latest-Ngapost2md ($Arch) {
    $api_gh = "https://api.github.com/repos/ludoux/ngapost2md/releases/latest"
    $json = Invoke-WebRequest $api_gh -MaximumRedirection 0 -ErrorAction Ignore -UseBasicParsing | ConvertFrom-Json
    $filename = $json.assets | Where-Object { $_.name -Match "windows-$Arch" } | Select-Object -ExpandProperty name
    $size = $json.assets | Where-Object { $_.name -Match "windows-$Arch" } | Select-Object -ExpandProperty size
    $title = $json.name
    $date = $json.created_at
    $body = $json.body
    $download_link = $json.assets | Where-Object { $_.name -Match "windows-$Arch" } | Select-Object -ExpandProperty browser_download_url
    if ($filename -is [array]) {
        return $title, $filename[0], $download_link[0]
    }
    else {
        return $title, $date, $body, $filename, $download_link, $size
    }
}

function Check-Local-Version() {
    $FilePath = [System.IO.Path]::Combine((Get-Location).Path, 'ngapost2md.exe')
    if (Test-Path $FilePath) {
        $output = cmd /c $FilePath "--version" 2`>`&1
        $ver = $output -split " "
        return $ver[1]
    } else {
        Write-Host "No ngapost2md found!"
        return "not_found"
    }
    
}

function Check-OS-Arch() {
    if (Test-Path (Join-Path $env:windir "SysWow64")) {
        $original_arch = "amd64"
    }
    else {
        $original_arch = "386"
    }
    return $original_arch
}

function Check-And-Download-And-Unzip() {
    $local_version = Check-Local-Version
    $arch = Check-OS-Arch

    Write-Host "You are using ngapost2md" $local_version ". Start checking latest version from GitHub release page..."

    # get release info
    $r_title, $r_date, $r_body, $r_filename, $r_link, $r_filesize = Get-Latest-Ngapost2md $arch
    
    # check if using the latest version
    if ($r_title -cmatch ("\[",$local_version,"\]" -join "")) {
        if (Test-Path $update_download_dir) {
            Remove-Item -Path $update_download_dir -Recurse -ErrorAction Stop
        }
        Write-Host ("Congratulations! You are using the latest version of ngapost2md:", $local_version) -ForegroundColor White -BackgroundColor Green
        Write-Host "Operation completed" -ForegroundColor Magenta
        timeout 8
        exit 0
    }

    # It may happen when the github actions do not release this arch
    if (!$r_filename) {
        Write-Host ("Can not find release binary for windows-", $Arch -join "") -ForegroundColor Red
        Write-Host "Operation completed" -ForegroundColor Magenta
        timeout 8
        exit 1
    }
    
    # print new version info
    Write-Host "New version (" $r_title ") found. Released at" $r_date
    Write-Host  "You can view changelog at https://github.com/ludoux/ngapost2md/releases/latest"
    
    # download it
    Download-Archive $r_filename $r_link

    # check file size
    $downloaded_file = Join-Path $update_download_dir $r_filename
    $file_size = (Get-ChildItem $downloaded_file).Length
    if ($file_size -eq $r_filesize) {
        Write-Host "Download completed. The archive file size is" $file_size
    } else {
        Write-Host "Downloaded file size is not correct! Please re-run this update script to redo!" -ForegroundColor Red
        Write-Host "Operation completed" -ForegroundColor Magenta
        timeout 8
        exit 1
    }

    # unzip file
    Expand-Archive $downloaded_file -DestinationPath $update_download_dir
}

function Update-Ps1-If-Necessary() {
    #May contains newer ps1 files. Update myself necessary
    if (Test-Path (Join-Path $update_download_dir $ps1_filename)) {
        $my_hash = (Get-FileHash $PSCommandPath).hash
        $it_hash = (Get-FileHash (Join-Path $update_download_dir $ps1_filename)).hash
        if ($my_hash -ne $it_hash) {
            Write-Host "Extract" $ps1_filename "..."
            Get-Content (Join-Path $update_download_dir $ps1_filename) -Raw | Set-Content $PSCommandPath
            return $true
        }
    }
    return $false
}

function Do-Update-After-Ps1-No-Change() {
    #Copy README.md
    Write-Host "Overwrite README.md ..."
    Override-Copy (Join-Path $update_download_dir "\README.md") (Get-Location).Path

    #Copy ngapost2md.exe
    Write-Host "Overwrite ngapost2md.exe ..."
    Override-Copy (Join-Path $update_download_dir "\ngapost2md.exe") (Get-Location).Path

    #Copy LICENSE
    Write-Host "Overwrite LICENSE ..."
    Override-Copy (Join-Path $update_download_dir "\LICENSE") (Get-Location).Path

    #Copy config.ini if it does not exist
    if (Test-Path (Join-Path (Get-Location) "\config.ini")) {
        Write-Host "Skip modify config.ini"
    } else {
        Write-Host "Write config.ini ..."
        Override-Copy (Join-Path $update_download_dir "\config.ini") (Get-Location).Path
    }
}

function Do-Update-After-Ps1-Change() {
    Do-Update-After-Ps1-No-Change
}

#
# Main script entry point
#
try {
    if (($args.Count -eq 1) -and ($args[0] -eq "continue_after_ps1_update")) {
        Do-Update-After-Ps1-Change
    } else {
        # remove temp dir at first and recreate it
        if (Test-Path $update_download_dir) {
            Remove-Item -Path $update_download_dir -Recurse -ErrorAction Stop
        }
        # create empty download dir
        # hide output
        $null = New-Item -ItemType Directory -Force $update_download_dir
        
        Check-And-Download-And-Unzip

        if (Update-Ps1-If-Necessary -eq $true) {
            # Re-run myself to continue updating
            & $PSCommandPath "continue_after_ps1_update"
            exit 0
        } else {
            Do-Update-After-Ps1-No-Change
        }
    }
    
    if (Test-Path $update_download_dir) {
        Remove-Item -Path $update_download_dir -Recurse -ErrorAction Stop
    }
    Write-Host "Update finished successfully!" -ForegroundColor White -BackgroundColor Green
    Write-Host "Operation completed" -ForegroundColor Magenta
    timeout 8
    exit 0
}
catch {
    if (Test-Path $update_download_dir) {
        Remove-Item -Path $update_download_dir -Recurse -ErrorAction Stop
    }
    Write-Host $_.Exception.Message -ForegroundColor Red
    timeout 8
    exit 1
}