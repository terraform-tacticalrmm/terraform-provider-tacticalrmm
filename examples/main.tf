terraform {
  required_providers {
    tacticalrmm = {
      source = "terraform-tacticalrmm/tacticalrmm"
      version = "~> 0.1.0"
    }
  }
}

provider "tacticalrmm" {
  endpoint = "https://api.your-trmm-instance.com"
  api_key  = var.trmm_api_key
}

# Example: Create a PowerShell script
resource "tacticalrmm_script" "disk_cleanup" {
  name        = "Disk Cleanup Script"
  description = "Cleans up temporary files and old logs"
  shell       = "powershell"
  category    = "Maintenance"
  
  script_body = <<-EOT
    # Clean Windows temp files
    Remove-Item -Path "$env:TEMP\*" -Force -Recurse -ErrorAction SilentlyContinue
    
    # Clean old log files
    Get-ChildItem -Path "C:\Logs" -Filter "*.log" | 
      Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-30) } | 
      Remove-Item -Force
    
    Write-Output "Cleanup completed successfully"
  EOT
  
  default_timeout = 300
  run_as_user     = false
  
  args = [
    "-ExecutionPolicy", "Bypass"
  ]
  
  supported_platforms = ["windows"]
}

# Example: Create a script snippet that can be reused
resource "tacticalrmm_script_snippet" "get_disk_space" {
  name  = "GetDiskSpace"
  desc  = "Gets available disk space"
  shell = "powershell"
  
  code = <<-EOT
    Get-PSDrive -PSProvider FileSystem | 
      Select-Object Name, @{N='FreeGB';E={[math]::Round($_.Free/1GB,2)}}, 
                    @{N='UsedGB';E={[math]::Round($_.Used/1GB,2)}}, 
                    @{N='TotalGB';E={[math]::Round(($_.Free+$_.Used)/1GB,2)}}
  EOT
}

# Example: Create a script that uses the snippet
resource "tacticalrmm_script" "monitor_disk" {
  name        = "Monitor Disk Space"
  description = "Monitors disk space and alerts if low"
  shell       = "powershell"
  category    = "Monitoring"
  
  script_body = <<-EOT
    # Get disk space using snippet
    {{GetDiskSpace}}
    
    # Check if any drive has less than 10% free space
    $drives = Get-PSDrive -PSProvider FileSystem
    foreach ($drive in $drives) {
      $percentFree = ($drive.Free / ($drive.Free + $drive.Used)) * 100
      if ($percentFree -lt 10) {
        Write-Warning "Drive $($drive.Name): has only $([math]::Round($percentFree,2))% free space!"
        exit 1
      }
    }
    
    Write-Output "All drives have sufficient free space"
    exit 0
  EOT
  
  default_timeout = 60
}

# Example: Store sensitive configuration in keystore
resource "tacticalrmm_keystore" "smtp_password" {
  name  = "smtp_password"
  value = var.smtp_password
}

resource "tacticalrmm_keystore" "api_endpoint" {
  name  = "external_api_endpoint"
  value = "https://api.external-service.com/v1"
}

# Example: Python script
resource "tacticalrmm_script" "python_example" {
  name        = "System Info Collector"
  description = "Collects system information"
  shell       = "python"
  category    = "Information"
  
  script_body = <<-EOT
    import platform
    import psutil
    import json
    
    info = {
        "platform": platform.platform(),
        "processor": platform.processor(),
        "cpu_count": psutil.cpu_count(),
        "memory_total": psutil.virtual_memory().total,
        "memory_available": psutil.virtual_memory().available,
        "disk_usage": []
    }
    
    for partition in psutil.disk_partitions():
        try:
            usage = psutil.disk_usage(partition.mountpoint)
            info["disk_usage"].append({
                "device": partition.device,
                "mountpoint": partition.mountpoint,
                "total": usage.total,
                "used": usage.used,
                "free": usage.free,
                "percent": usage.percent
            })
        except:
            pass
    
    print(json.dumps(info, indent=2))
  EOT
  
  default_timeout     = 30
  supported_platforms = ["windows", "linux", "darwin"]
}

# Variables
variable "trmm_api_key" {
  description = "Tactical RMM API Key"
  type        = string
  sensitive   = true
}

variable "smtp_password" {
  description = "SMTP Password for email notifications"
  type        = string
  sensitive   = true
}
