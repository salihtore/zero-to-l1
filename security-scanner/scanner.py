#!/usr/bin/env python3
import sys
import os
import argparse

# Force UTF-8 stdout and stderr encoding on Windows
if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')
if hasattr(sys.stderr, 'reconfigure'):
    sys.stderr.reconfigure(encoding='utf-8')

# Ensure Python Scripts directory is on PATH for solc discovery
python_scripts = os.path.expanduser(r"~\AppData\Roaming\Python\Python310\Scripts")
if os.path.exists(python_scripts) and python_scripts not in os.environ.get("PATH", ""):
    os.environ["PATH"] = python_scripts + os.path.pathsep + os.environ.get("PATH", "")

from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.text import Text

from slither import Slither
from detectors import CUSTOM_DETECTORS

console = Console()

def run_scanner(target_path: str):
    if not os.path.exists(target_path):
        console.print(f"[bold red]❌ Error:[/bold red] Target path '{target_path}' does not exist.")
        sys.exit(1)

    console.print(Panel(f"[bold cyan]🔍 Zero to Secure L1 - Security Scanner (Slither MVP)[/bold cyan]\n[yellow]Target:[/yellow] {target_path}", expand=False))

    try:
        slither = Slither(target_path)
    except Exception as e:
        console.print(f"[bold red]❌ Slither Compilation Error:[/bold red] {e}")
        sys.exit(1)

    # Register custom detectors
    for detector_cls in CUSTOM_DETECTORS:
        slither.register_detector(detector_cls)

    # Run custom detectors
    raw_results = slither.run_detectors()

    findings = []
    for detector_result in raw_results:
        for res in detector_result:
            # Check if this finding belongs to one of our custom detectors
            check_name = res.get("check", "")
            description = res.get("description", "")
            elements = res.get("elements", [])

            file_info = "Unknown"
            line_info = "-"

            if elements:
                first_elem = elements[0]
                source_mapping = first_elem.get("source_mapping", {})
                if source_mapping:
                    filename_relative = source_mapping.get("filename_relative", "")
                    lines = source_mapping.get("lines", [])
                    if filename_relative:
                        file_info = os.path.basename(filename_relative)
                    if lines:
                        line_info = f"L{lines[0]}"

            findings.append({
                "check": check_name,
                "file": file_info,
                "line": line_info,
                "description": description.strip(),
            })

    # Render results with rich formatting
    if not findings:
        console.print("\n[bold green]✅ CONGRATULATIONS! No Teleporter/ICM security vulnerabilities found.[/bold green]\n")
        return 0

    table = Table(title="⚠️  Security Vulnerabilities Detected", show_header=True, header_style="bold magenta")
    table.add_column("Status", style="bold red", justify="center", width=8)
    table.add_column("Detector", style="bold yellow", width=28)
    table.add_column("Location", style="cyan", width=25)
    table.add_column("Description & Impact", style="white")

    for f in findings:
        table.add_row(
            "❌",
            f["check"],
            f"{f['file']}:{f['line']}",
            f["description"]
        )

    console.print()
    console.print(table)
    console.print(f"\n[bold red]❌ Total Findings: {len(findings)}[/bold red]\n")
    return len(findings)

def main():
    parser = argparse.ArgumentParser(description="Zero to Secure L1 - Avalanche Teleporter Security Scanner")
    parser.add_argument("--target", required=True, help="Solidity file or directory path to scan")
    args = parser.parse_args()

    findings_count = run_scanner(args.target)
    sys.exit(0 if findings_count == 0 else 1)

if __name__ == "__main__":
    main()
