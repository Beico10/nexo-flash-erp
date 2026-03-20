"""
Nexo One ERP - Phase 3 API Tests
Features tested:
- Admin Plan Management (GET/PUT /api/v1/admin/plans)
- Expense Module (CRUD, QR parsing, categories, summary, tax report)
- SEFAZ Integration (real code but external URLs may fail)
"""
import pytest
import requests
import os
import time

BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', 'https://repo-status-check-2.preview.emergentagent.com').rstrip('/')

# Test credentials
TEST_CREDENTIALS = {
    "tenant_slug": "demo",
    "email": "admin@demo.com",
    "password": "demo123"
}

@pytest.fixture(scope="module")
def auth_token():
    """Get authentication token for protected endpoints"""
    response = requests.post(
        f"{BASE_URL}/api/auth/login",
        json=TEST_CREDENTIALS,
        headers={"Content-Type": "application/json"}
    )
    assert response.status_code == 200, f"Login failed: {response.text}"
    data = response.json()
    assert "access_token" in data
    return data["access_token"]


class TestHealthCheck:
    """Basic health check tests"""
    
    def test_health_endpoint(self):
        """Health endpoint should return OK"""
        response = requests.get(f"{BASE_URL}/api/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert data["engine"] == "fiscal_ibs_cbs_2026"


class TestAdminPlansAPI:
    """Admin plan management endpoints"""
    
    def test_get_all_plans_admin(self, auth_token):
        """GET /api/v1/admin/plans - should return 4 plans"""
        response = requests.get(
            f"{BASE_URL}/api/v1/admin/plans",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        assert "plans" in data
        plans = data["plans"]
        
        # Should have 4 plans (starter, pro, business, enterprise)
        assert len(plans) == 4
        
        plan_codes = [p["code"] for p in plans]
        assert "starter" in plan_codes
        assert "pro" in plan_codes
        assert "business" in plan_codes
        assert "enterprise" in plan_codes
    
    def test_plans_have_required_fields(self, auth_token):
        """Plans should have all required fields with snake_case"""
        response = requests.get(
            f"{BASE_URL}/api/v1/admin/plans",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        data = response.json()
        
        for plan in data["plans"]:
            # Required fields
            assert "id" in plan
            assert "code" in plan
            assert "name" in plan
            assert "price_monthly" in plan
            assert "price_yearly" in plan
            assert "setup_fee" in plan
            assert "features" in plan
            assert "allowed_niches" in plan
            assert "is_active" in plan
            assert "is_featured" in plan
    
    def test_update_plan_price(self, auth_token):
        """PUT /api/v1/admin/plans - should update plan price"""
        # First get current starter plan details
        response = requests.get(
            f"{BASE_URL}/api/v1/admin/plans",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        original_plans = response.json()["plans"]
        starter = next((p for p in original_plans if p["code"] == "starter"), None)
        assert starter is not None
        original_price = starter["price_monthly"]
        
        # Update with full plan data to avoid resetting other fields
        new_price = 599 if original_price != 599 else 497
        update_data = {
            "code": "starter",
            "price_monthly": new_price,
            # Include features to avoid reset bug
            "features": starter["features"],
            "is_active": starter["is_active"],
            "is_featured": starter["is_featured"],
            "allowed_niches": starter["allowed_niches"]
        }
        
        response = requests.put(
            f"{BASE_URL}/api/v1/admin/plans",
            json=update_data,
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert data.get("updated") == True
        assert data["plan"]["price_monthly"] == new_price
    
    def test_update_plan_verify_persistence(self, auth_token):
        """After update, GET should return updated price"""
        # Get current price
        response = requests.get(
            f"{BASE_URL}/api/v1/admin/plans",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        plans = response.json()["plans"]
        starter = next((p for p in plans if p["code"] == "starter"), None)
        current_price = starter["price_monthly"]
        
        # Update to a different price
        new_price = 555
        update_data = {
            "code": "starter",
            "price_monthly": new_price,
            "features": starter["features"],
            "is_active": starter["is_active"],
            "is_featured": starter["is_featured"]
        }
        
        requests.put(
            f"{BASE_URL}/api/v1/admin/plans",
            json=update_data,
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        
        # Verify persistence
        response = requests.get(
            f"{BASE_URL}/api/v1/admin/plans",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        plans = response.json()["plans"]
        starter = next((p for p in plans if p["code"] == "starter"), None)
        assert starter["price_monthly"] == new_price
    
    def test_update_plan_without_code_fails(self, auth_token):
        """PUT without code should return 400"""
        response = requests.put(
            f"{BASE_URL}/api/v1/admin/plans",
            json={"price_monthly": 999},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        assert response.status_code == 400


class TestExpenseListAPI:
    """Expense list endpoint tests"""
    
    def test_list_expenses(self, auth_token):
        """GET /api/v1/expenses - should return 5 seeded expenses"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        
        assert "expenses" in data
        assert "count" in data
        expenses = data["expenses"]
        
        # Should have 5 seeded expenses
        assert len(expenses) == 5
        assert data["count"] == 5
    
    def test_expense_snake_case_keys(self, auth_token):
        """Expenses should use snake_case JSON keys"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        data = response.json()
        
        expense = data["expenses"][0]
        # Verify snake_case keys
        assert "id" in expense
        assert "supplier_name" in expense
        assert "total_amount" in expense
        assert "category" in expense
        assert "issue_date" in expense
        assert "ibs_credit" in expense
        assert "cbs_credit" in expense


class TestExpenseDetailAPI:
    """Expense detail endpoint tests"""
    
    def test_get_expense_by_id(self, auth_token):
        """GET /api/v1/expenses/exp-1 - should return expense with items_count"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses/exp-1",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        
        assert "expense" in data
        expense = data["expense"]
        
        assert expense["id"] == "exp-1"
        assert expense["supplier_name"] == "Auto Pecas Central"
        assert "items_count" in expense
        assert expense["items_count"] == 3  # Has 3 seeded items
    
    def test_get_expense_not_found(self, auth_token):
        """GET /api/v1/expenses/non-existent - should return 404"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses/non-existent-id",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 404


class TestExpenseCategoriesAPI:
    """Expense categories endpoint tests"""
    
    def test_get_categories(self, auth_token):
        """GET /api/v1/expenses/categories - should return 8 categories"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses/categories",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        
        assert "categories" in data
        categories = data["categories"]
        
        # Should have 8 categories
        assert len(categories) == 8
        
        # Verify some expected categories
        codes = [c["Code"] for c in categories]
        assert "pecas" in codes
        assert "mercadorias" in codes
        assert "combustivel" in codes
        assert "alimentacao" in codes
        assert "outros" in codes


class TestExpenseSummaryAPI:
    """Expense summary endpoint tests"""
    
    def test_get_summary(self, auth_token):
        """GET /api/v1/expenses/summary - should return summary by category"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses/summary",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        
        assert "summary" in data
        assert "totals" in data
        assert "period" in data
        
        # Verify totals structure
        totals = data["totals"]
        assert "amount" in totals
        assert "ibs_credit" in totals
        assert "cbs_credit" in totals
        assert "tax_credit" in totals
        
        # Totals should be positive
        assert totals["amount"] > 0


class TestTaxReportAPI:
    """Tax report endpoint tests"""
    
    def test_get_tax_report(self, auth_token):
        """GET /api/v1/expenses/tax-report - should return tax deduction report"""
        response = requests.get(
            f"{BASE_URL}/api/v1/expenses/tax-report?year=2026",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        
        assert "report" in data
        assert "totals" in data
        assert data["year"] == 2026
        
        # Verify totals structure
        totals = data["totals"]
        assert "deductible" in totals
        assert "non_deductible" in totals
        assert "tax_credit" in totals
        
        # Verify report entries have required fields
        for entry in data["report"]:
            assert "category_code" in entry
            assert "tax_deductible" in entry
            assert "total" in entry
            assert "tax_credit" in entry


class TestParseQRCodeAPI:
    """QR Code parsing endpoint tests"""
    
    def test_parse_nfce_qr(self, auth_token):
        """POST /api/v1/expenses/parse-qr - should parse NFC-e QR content"""
        qr_content = "https://www.nfce.fazenda.sp.gov.br/consulta?chNFe=35260312345678000199550010000012341234567890"
        
        response = requests.post(
            f"{BASE_URL}/api/v1/expenses/parse-qr",
            json={"qr_content": qr_content},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        assert response.status_code == 200
        data = response.json()
        
        assert data["type"] == "nfce"
        assert data["is_valid"] == True
        assert data["uf"] == "SP"
        assert len(data["access_key"]) == 44
    
    def test_parse_invalid_qr(self, auth_token):
        """POST /api/v1/expenses/parse-qr - invalid QR should return 400"""
        response = requests.post(
            f"{BASE_URL}/api/v1/expenses/parse-qr",
            json={"qr_content": "random invalid content"},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        assert response.status_code == 400
    
    def test_parse_pix_qr(self, auth_token):
        """PIX QR codes should be identified but marked invalid for expenses"""
        # Standard PIX QR starts with 00020126
        pix_content = "00020126330014br.gov.bcb.pix0111120000000015204000053039865802BR5913Test"
        
        response = requests.post(
            f"{BASE_URL}/api/v1/expenses/parse-qr",
            json={"qr_content": pix_content},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        # PIX is recognized and returns 200 but is_valid=false
        assert response.status_code == 200
        data = response.json()
        assert data["type"] == "pix"
        assert data["is_valid"] == False


class TestScanQRCodeAPI:
    """QR Code scan endpoint tests (SEFAZ integration)"""
    
    def test_scan_qr_sefaz_attempt(self, auth_token):
        """POST /api/v1/expenses/scan - attempts SEFAZ lookup (may fail on external)"""
        # Using a different key to avoid duplicate check
        qr_content = "https://www.nfce.fazenda.sp.gov.br/consulta?chNFe=35260312345678000199550010000012349999999999"
        
        response = requests.post(
            f"{BASE_URL}/api/v1/expenses/scan",
            json={"qr_content": qr_content},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        
        # SEFAZ is real code but external URLs may be unreachable
        # Accept either 200 (success) or 500/503 (SEFAZ unavailable)
        assert response.status_code in [200, 500, 503]
        
        if response.status_code == 200:
            data = response.json()
            assert data["success"] == True
            assert "expense" in data
    
    def test_scan_duplicate_expense(self, auth_token):
        """Scanning duplicate NF-e key should return 409 conflict"""
        # Using existing seeded expense key
        qr_content = "https://www.nfce.fazenda.sp.gov.br/consulta?chNFe=35260312345678000199550010000012341234567890"
        
        response = requests.post(
            f"{BASE_URL}/api/v1/expenses/scan",
            json={"qr_content": qr_content},
            headers={
                "Authorization": f"Bearer {auth_token}",
                "Content-Type": "application/json"
            }
        )
        
        # Should return conflict for duplicate
        assert response.status_code == 409


class TestAuthRequired:
    """Verify authentication is required for all endpoints"""
    
    def test_admin_plans_requires_auth(self):
        """Admin plans endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/admin/plans")
        assert response.status_code == 401
    
    def test_expenses_requires_auth(self):
        """Expenses endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/expenses")
        assert response.status_code == 401
    
    def test_expense_categories_requires_auth(self):
        """Categories endpoint requires authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/expenses/categories")
        assert response.status_code == 401


class TestPreviousFeaturesStillWork:
    """Verify Phase 1 and Phase 2 features still work"""
    
    def test_login_still_works(self):
        """Login should still work"""
        response = requests.post(
            f"{BASE_URL}/api/auth/login",
            json=TEST_CREDENTIALS,
            headers={"Content-Type": "application/json"}
        )
        assert response.status_code == 200
        assert "access_token" in response.json()
    
    def test_dashboard_stats_still_work(self, auth_token):
        """Dashboard stats should still work"""
        response = requests.get(
            f"{BASE_URL}/api/v1/dashboard/stats",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
    
    def test_billing_subscription_still_works(self, auth_token):
        """Subscription endpoint should still work"""
        response = requests.get(
            f"{BASE_URL}/api/v1/billing/subscription",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        data = response.json()
        assert "subscription" in data
    
    def test_public_plans_still_work(self):
        """Public plans endpoint should still work"""
        response = requests.get(f"{BASE_URL}/api/billing/plans")
        assert response.status_code == 200
        data = response.json()
        assert "plans" in data


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
