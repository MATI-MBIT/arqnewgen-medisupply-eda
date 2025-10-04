#!/usr/bin/env python3
"""
Test script for warehouse batch service
Provides build, test, and run functionality for local development
"""

import os
import sys
import subprocess
import argparse
import json
from pathlib import Path

class WarehouseBatchTester:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.bin_dir = self.project_root / "bin"
        self.src_dir = self.project_root / "src"
        self.examples_dir = self.project_root / "examples"
        
        # Default environment variables
        self.env_vars = {
            "KAFKA_ORDER_EVENTS_TOPIC": "order-events", 
            "KAFKA_BROKER_ADDRESS": "localhost:9092",
            "KAFKA_GROUP_ID": "warehouse-batch-service",
            "HTTP_PORT": "8080"
        }
    
    def setup_environment(self):
        """Set up environment variables"""
        print("Setting up environment variables...")
        for key, value in self.env_vars.items():
            os.environ[key] = value
            print(f"  {key}={value}")
    
    def create_bin_dir(self):
        """Create bin directory if it doesn't exist"""
        self.bin_dir.mkdir(exist_ok=True)
    
    def build(self):
        """Build the Go application"""
        print("\n🔨 Building application...")
        self.create_bin_dir()
        
        cmd = ["go", "build", "-o", str(self.bin_dir / "warehouse-batch"), "src/main.go"]
        result = subprocess.run(cmd, cwd=self.project_root, capture_output=True, text=True)
        
        if result.returncode == 0:
            print("✅ Build successful!")
            return True
        else:
            print("❌ Build failed!")
            print(result.stderr)
            return False
    
    def test(self):
        """Run Go tests"""
        print("\n🧪 Running tests...")
        
        cmd = ["go", "test", "./src/application/", "-v"]
        result = subprocess.run(cmd, cwd=self.project_root, capture_output=True, text=True)
        
        if result.returncode == 0:
            print("✅ All tests passed!")
            print(result.stdout)
            return True
        else:
            print("❌ Tests failed!")
            print(result.stderr)
            return False
    
    def validate_examples(self):
        """Validate example JSON files"""
        print("\n📋 Validating example JSON files...")
        
        json_files = list(self.examples_dir.glob("*.json"))
        if not json_files:
            print("⚠️  No example JSON files found")
            return True
        
        all_valid = True
        for json_file in json_files:
            try:
                with open(json_file, 'r') as f:
                    data = json.load(f)
                print(f"✅ {json_file.name} - Valid JSON")
                
                # Validate order event structure
                if "event_type" in data and "order_id" in data and "order" in data:
                    print(f"   Event Type: {data['event_type']}")
                    print(f"   Order ID: {data['order_id']}")
                else:
                    print(f"⚠️  {json_file.name} - Missing required fields")
                    
            except json.JSONDecodeError as e:
                print(f"❌ {json_file.name} - Invalid JSON: {e}")
                all_valid = False
            except Exception as e:
                print(f"❌ {json_file.name} - Error: {e}")
                all_valid = False
        
        return all_valid
    
    def run_service(self):
        """Run the warehouse batch service"""
        print("\n🚀 Starting warehouse batch service...")
        print("Press Ctrl+C to stop the service")
        
        binary_path = self.bin_dir / "warehouse-batch"
        if not binary_path.exists():
            print("❌ Binary not found. Please run build first.")
            return False
        
        try:
            subprocess.run([str(binary_path)], cwd=self.project_root, env=os.environ)
        except KeyboardInterrupt:
            print("\n🛑 Service stopped by user")
        except Exception as e:
            print(f"❌ Error running service: {e}")
            return False
        
        return True
    
    def clean(self):
        """Clean build artifacts"""
        print("\n🧹 Cleaning build artifacts...")
        
        if self.bin_dir.exists():
            for file in self.bin_dir.iterdir():
                if file.is_file():
                    file.unlink()
                    print(f"  Removed {file.name}")
        
        print("✅ Clean completed!")
    
    def check_dependencies(self):
        """Check if required dependencies are available"""
        print("\n🔍 Checking dependencies...")
        
        # Check Go
        try:
            result = subprocess.run(["go", "version"], capture_output=True, text=True)
            if result.returncode == 0:
                print(f"✅ Go: {result.stdout.strip()}")
            else:
                print("❌ Go not found")
                return False
        except FileNotFoundError:
            print("❌ Go not found")
            return False
        
        # Check if go.mod exists
        go_mod = self.project_root / "go.mod"
        if go_mod.exists():
            print("✅ go.mod found")
        else:
            print("❌ go.mod not found")
            return False
        
        return True
    
    def show_config(self):
        """Show current configuration"""
        print("\n⚙️  Current Configuration:")
        for key, value in self.env_vars.items():
            current_value = os.environ.get(key, value)
            print(f"  {key}={current_value}")
        
        print(f"\nProject Root: {self.project_root}")
        print(f"Binary Path: {self.bin_dir / 'warehouse-batch'}")

def main():
    parser = argparse.ArgumentParser(description="Warehouse Batch Service Test Tool")
    parser.add_argument("command", nargs="?", default="all", 
                       choices=["all", "build", "test", "run", "clean", "deps", "config", "validate"],
                       help="Command to execute")
    parser.add_argument("--env", action="store_true", help="Load environment from .env file")
    
    args = parser.parse_args()
    
    tester = WarehouseBatchTester()
    
    # Load .env file if requested
    if args.env:
        env_file = tester.project_root / ".env"
        if env_file.exists():
            print(f"📁 Loading environment from {env_file}")
            with open(env_file) as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith("#") and "=" in line:
                        key, value = line.split("=", 1)
                        tester.env_vars[key] = value
    
    tester.setup_environment()
    
    success = True
    
    if args.command == "all":
        print("🎯 Running full test suite...")
        success &= tester.check_dependencies()
        success &= tester.validate_examples()
        success &= tester.build()
        success &= tester.test()
        print("\n🎉 Full test suite completed!" if success else "\n💥 Some tests failed!")
        
    elif args.command == "build":
        success = tester.build()
        
    elif args.command == "test":
        success = tester.test()
        
    elif args.command == "run":
        if not tester.build():
            sys.exit(1)
        tester.run_service()
        
    elif args.command == "clean":
        tester.clean()
        
    elif args.command == "deps":
        success = tester.check_dependencies()
        
    elif args.command == "config":
        tester.show_config()
        
    elif args.command == "validate":
        success = tester.validate_examples()
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()