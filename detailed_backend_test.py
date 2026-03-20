#!/usr/bin/env python3
"""
Extended Backend API Tests - Verification of specific review requirements
"""

import requests
import sys
import json
from datetime import datetime

class DetailedNexoAPITester:
    def __init__(self, base_url: str = "https://repo-status-check-2.preview.emergentagent.com"):
        self.base_url = base_url
        self.access_token = None
        self.session = requests.Session()
        
    def authenticate(self):
        """Get access token for testing protected endpoints"""
        login_data = {
            "tenant_slug": "demo",
            "email": "admin@demo.com", 
            "password": "demo123"
        }
        
        response = self.session.post(
            f"{self.base_url}/api/auth/login",
            json=login_data,
            headers={'Content-Type': 'application/json'}
        )
        
        if response.status_code == 200:
            data = response.json()
            self.access_token = data.get('access_token')
            return True
        return False

    def test_health_detailed(self):
        """Detailed health check verification"""
        print("🔍 Testing Health Check API in Detail...")
        
        response = self.session.get(f"{self.base_url}/api/health")
        
        if response.status_code == 200:
            data = response.json()
            
            print(f"  ✅ Status Code: 200")
            print(f"  📊 Full Response: {json.dumps(data, indent=2)}")
            
            # Verify specific requirements from review request
            checks = [
                ("status", "ok", data.get('status')),
                ("stack", "go-puro", data.get('stack')), 
                ("version", "contains version", data.get('version')),
                ("engine", "fiscal_ibs_cbs_2026", data.get('engine'))
            ]
            
            all_passed = True
            for field, expected, actual in checks:
                if field == "version":
                    passed = actual is not None and len(str(actual)) > 0
                else:
                    passed = actual == expected
                
                status = "✅" if passed else "❌"
                print(f"  {status} {field}: expected '{expected}', got '{actual}'")
                if not passed:
                    all_passed = False
            
            return all_passed
        else:
            print(f"  ❌ Health check failed with status {response.status_code}")
            return False

    def test_dashboard_stats_detailed(self):
        """Detailed dashboard stats verification"""
        if not self.access_token:
            print("❌ Cannot test dashboard stats - no access token")
            return False
            
        print("🔍 Testing Dashboard Stats API in Detail...")
        
        headers = {'Authorization': f'Bearer {self.access_token}'}
        response = self.session.get(f"{self.base_url}/api/v1/dashboard/stats", headers=headers)
        
        if response.status_code == 200:
            data = response.json()
            print(f"  ✅ Status Code: 200")
            print(f"  📊 Dashboard Data Structure:")
            
            # Verify expected data structure
            expected_keys = ['revenue', 'mechanic_os', 'appointments', 'bakery_products', 'pending_suggestions']
            
            for key in expected_keys:
                if key in data:
                    value = data[key]
                    print(f"    ✅ {key}: {type(value).__name__} - {value}")
                else:
                    print(f"    ❌ {key}: MISSING")
            
            # Check revenue structure if present
            if 'revenue' in data and isinstance(data['revenue'], dict):
                revenue = data['revenue']
                print(f"    📈 Revenue details:")
                for k, v in revenue.items():
                    print(f"      - {k}: {v}")
            
            return True
        else:
            print(f"  ❌ Dashboard stats failed with status {response.status_code}")
            print(f"  📝 Response: {response.text}")
            return False

    def test_ai_copilot_detailed(self):
        """Detailed AI Co-Piloto verification with IBS/CBS question"""
        if not self.access_token:
            print("❌ Cannot test AI Co-Piloto - no access token")
            return False
            
        print("🔍 Testing AI Co-Piloto API in Detail...")
        
        # IBS/CBS specific question as mentioned in review
        ai_data = {
            "question": "Explique o impacto da transição do modelo tributário atual para IBS e CBS em 2026. Como isso afeta pequenas empresas?",
            "context": "Sistema ERP Nexo One - consultoria fiscal"
        }
        
        headers = {
            'Authorization': f'Bearer {self.access_token}',
            'Content-Type': 'application/json'
        }
        
        response = self.session.post(
            f"{self.base_url}/api/v1/copilot/suggest", 
            json=ai_data, 
            headers=headers
        )
        
        if response.status_code == 200:
            data = response.json()
            print(f"  ✅ Status Code: 200")
            
            suggestion = data.get('suggestion', '')
            model = data.get('model', '')
            session_id = data.get('session_id', '')
            
            print(f"  🤖 AI Model: {model}")
            print(f"  🆔 Session ID: {session_id}")
            print(f"  💬 Response Length: {len(suggestion)} characters")
            
            if len(suggestion) > 50:
                print(f"  📝 Response Preview: {suggestion[:200]}...")
                
                # Check if response mentions IBS/CBS
                ibs_cbs_mentioned = any(term in suggestion.lower() for term in ['ibs', 'cbs', 'reforma tributária', 'tribut'])
                print(f"  🎯 IBS/CBS Context: {'✅ Mentioned' if ibs_cbs_mentioned else '❌ Not mentioned'}")
                
                return True
            else:
                print(f"  ❌ Response too short or empty")
                return False
        else:
            print(f"  ❌ AI Co-Piloto failed with status {response.status_code}")
            print(f"  📝 Response: {response.text}")
            return False

    def run_detailed_tests(self):
        """Run all detailed tests"""
        print("🚀 Starting Detailed Backend API Tests")
        print("=" * 60)
        
        results = []
        
        # Health check
        health_result = self.test_health_detailed()
        results.append(("Health Check API", health_result))
        
        print()
        
        # Authentication
        auth_result = self.authenticate()
        if auth_result:
            print("✅ Authentication successful")
        else:
            print("❌ Authentication failed")
        results.append(("Authentication", auth_result))
        
        print()
        
        # Dashboard stats
        dashboard_result = self.test_dashboard_stats_detailed()
        results.append(("Dashboard Stats API", dashboard_result))
        
        print()
        
        # AI Co-Piloto
        ai_result = self.test_ai_copilot_detailed()
        results.append(("AI Co-Piloto API", ai_result))
        
        # Summary
        print("\n" + "=" * 60)
        print("📋 DETAILED TEST RESULTS:")
        
        passed = 0
        total = len(results)
        
        for test_name, result in results:
            status = "✅ PASS" if result else "❌ FAIL"
            print(f"  {status} {test_name}")
            if result:
                passed += 1
        
        print(f"\n🏆 Overall: {passed}/{total} tests passed ({(passed/total)*100:.1f}%)")
        
        return 0 if passed == total else 1

def main():
    tester = DetailedNexoAPITester()
    return tester.run_detailed_tests()

if __name__ == "__main__":
    sys.exit(main())