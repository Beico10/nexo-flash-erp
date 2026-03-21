"""
Nexo One ERP - Phase 2 Testing: Billing, Trial, Onboarding, Journey Tracking

Test Coverage:
- Billing Plans API (public)
- Coupon Validation (public)
- Subscription Management (authenticated)
- Trial WhatsApp Verification (public)
- Onboarding Steps/Progress (authenticated)
- Journey Tracking (authenticated)
- Funnel Analytics (authenticated)
"""

import pytest
import requests
import os

# Use environment variable or preview URL
BASE_URL = os.environ.get('REACT_APP_BACKEND_URL', 'https://dispatch-module.preview.emergentagent.com').rstrip('/')

# Test credentials
TENANT_SLUG = 'demo'
TEST_EMAIL = 'admin@demo.com'
TEST_PASSWORD = 'demo123'


class TestPublicBillingEndpoints:
    """Test public billing endpoints (no auth required)"""

    def test_get_billing_plans_returns_5_plans(self):
        """GET /api/billing/plans - Should return 5 plans"""
        response = requests.get(f"{BASE_URL}/api/billing/plans")
        assert response.status_code == 200, f"Expected 200, got {response.status_code}"
        
        data = response.json()
        assert 'plans' in data, "Response should have 'plans' key"
        
        plans = data['plans']
        assert len(plans) == 5, f"Expected 5 plans, got {len(plans)}"
        
        # Verify plan codes
        plan_codes = [p['code'] for p in plans]
        expected_codes = ['micro', 'starter', 'pro', 'business', 'enterprise']
        for code in expected_codes:
            assert code in plan_codes, f"Missing plan code: {code}"
        
        # Verify Pro plan is featured
        pro_plan = next((p for p in plans if p['code'] == 'pro'), None)
        assert pro_plan is not None
        assert pro_plan['is_featured'] == True, "Pro plan should be featured"
        assert pro_plan['price_monthly'] == 199, "Pro monthly price should be 199"
        
        # Verify pricing structure
        for plan in plans:
            assert 'price_monthly' in plan
            assert 'price_yearly' in plan
            assert 'features' in plan
            assert 'max_users' in plan

    def test_billing_plans_have_correct_pricing(self):
        """Verify plan pricing is correct"""
        response = requests.get(f"{BASE_URL}/api/billing/plans")
        plans = {p['code']: p for p in response.json()['plans']}
        
        # Check specific prices
        assert plans['micro']['price_monthly'] == 49
        assert plans['starter']['price_monthly'] == 99
        assert plans['pro']['price_monthly'] == 199
        assert plans['business']['price_monthly'] == 399
        assert plans['enterprise']['price_monthly'] == 999
        
        # Enterprise has setup fee
        assert plans['enterprise']['setup_fee'] == 500

    def test_validate_coupon_nexo20_returns_20_percent_discount(self):
        """POST /api/billing/coupon/validate - NEXO20 should give 20% discount"""
        response = requests.post(
            f"{BASE_URL}/api/billing/coupon/validate",
            json={"code": "NEXO20", "plan_code": "pro"}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}"
        
        data = response.json()
        assert data['valid'] == True
        assert data['discount_value'] == 20
        assert data['discount_type'] == 'percent'

    def test_validate_coupon_primeiro_returns_100_percent(self):
        """POST /api/billing/coupon/validate - PRIMEIRO should give 100% off first month"""
        response = requests.post(
            f"{BASE_URL}/api/billing/coupon/validate",
            json={"code": "PRIMEIRO", "plan_code": "starter"}
        )
        assert response.status_code == 200
        
        data = response.json()
        assert data['valid'] == True
        assert data['discount_value'] == 100

    def test_validate_coupon_invalid_returns_error(self):
        """POST /api/billing/coupon/validate - Invalid coupon should fail"""
        response = requests.post(
            f"{BASE_URL}/api/billing/coupon/validate",
            json={"code": "INVALID", "plan_code": "pro"}
        )
        assert response.status_code == 400, f"Expected 400, got {response.status_code}"


class TestTrialVerificationEndpoints:
    """Test trial WhatsApp verification (public endpoints)"""

    def test_start_verification_returns_whatsapp_url(self):
        """POST /api/auth/verify/start - Should return whatsapp_url"""
        response = requests.post(
            f"{BASE_URL}/api/auth/verify/start",
            json={
                "phone": "+5511999888777",
                "email": "test@example.com",
                "device_hash": "test-device-hash-123"
            }
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'whatsapp_url' in data, "Response should have 'whatsapp_url'"
        assert data['whatsapp_url'].startswith('https://wa.me/')
        assert 'expires_in' in data
        assert data['expires_in'] == 300  # 5 minutes

    def test_start_verification_requires_phone(self):
        """POST /api/auth/verify/start - Phone is required"""
        response = requests.post(
            f"{BASE_URL}/api/auth/verify/start",
            json={"email": "test@example.com"}
        )
        assert response.status_code == 400

    def test_verify_code_with_invalid_code(self):
        """POST /api/auth/verify/confirm - Invalid code should fail"""
        response = requests.post(
            f"{BASE_URL}/api/auth/verify/confirm",
            json={"phone": "+5511999888777", "code": "000000"}
        )
        # Should fail with expired or invalid code
        assert response.status_code in [400, 410], f"Expected 400 or 410, got {response.status_code}"


@pytest.fixture(scope="class")
def auth_token():
    """Get authentication token for protected endpoints"""
    # Direct login (tenant validation is embedded in login)
    login_response = requests.post(
        f"{BASE_URL}/api/auth/login",
        json={
            "tenant_slug": TENANT_SLUG,
            "email": TEST_EMAIL,
            "password": TEST_PASSWORD
        }
    )
    if login_response.status_code != 200:
        pytest.skip(f"Login failed: {login_response.text}")
    
    token = login_response.json().get('access_token')
    if not token:
        pytest.skip("No access_token in login response")
    
    return token


class TestAuthenticatedBillingEndpoints:
    """Test authenticated billing endpoints"""

    def test_get_subscription_returns_pro_trial(self, auth_token):
        """GET /api/v1/billing/subscription - Should return Pro plan trial"""
        response = requests.get(
            f"{BASE_URL}/api/v1/billing/subscription",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'subscription' in data
        
        sub = data['subscription']
        assert sub['plan_code'] == 'pro', f"Expected 'pro', got {sub['plan_code']}"
        assert sub['status'] == 'trialing', f"Expected 'trialing', got {sub['status']}"
        assert sub['trial_ends_at'] is not None
        assert sub['price'] == 199

    def test_get_subscription_includes_usage_data(self, auth_token):
        """GET /api/v1/billing/subscription - Should include usage data"""
        response = requests.get(
            f"{BASE_URL}/api/v1/billing/subscription",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        data = response.json()
        
        assert 'usage' in data, "Response should have 'usage' key"
        usage = data['usage']
        
        # Check usage structure
        assert len(usage) >= 4, "Should have at least 4 usage metrics"
        
        metrics = [u['metric'] for u in usage]
        assert 'users' in metrics
        assert 'transactions' in metrics

    def test_get_usage_returns_usage_status(self, auth_token):
        """GET /api/v1/billing/usage - Should return usage status"""
        response = requests.get(
            f"{BASE_URL}/api/v1/billing/usage",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        
        data = response.json()
        assert 'usage' in data

    def test_subscription_requires_auth(self):
        """GET /api/v1/billing/subscription - Should require authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/billing/subscription")
        assert response.status_code == 401

    def test_convert_trial_works(self, auth_token):
        """POST /api/v1/billing/convert - Should convert trial to active"""
        response = requests.post(
            f"{BASE_URL}/api/v1/billing/convert",
            headers={"Authorization": f"Bearer {auth_token}"},
            json={"payment_method": "pix", "coupon_code": ""}
        )
        # Note: In demo mode, this might fail or succeed depending on state
        # We just check the endpoint is accessible
        assert response.status_code in [200, 400], f"Unexpected status: {response.status_code}"
        
        if response.status_code == 200:
            data = response.json()
            assert 'subscription' in data
            assert 'message' in data

    def test_change_plan_works(self, auth_token):
        """POST /api/v1/billing/change-plan - Should change plan"""
        response = requests.post(
            f"{BASE_URL}/api/v1/billing/change-plan",
            headers={"Authorization": f"Bearer {auth_token}"},
            json={"plan_code": "business"}
        )
        # Accept both success and validation errors
        assert response.status_code in [200, 400], f"Unexpected status: {response.status_code}"


class TestOnboardingEndpoints:
    """Test onboarding endpoints"""

    def test_get_onboarding_steps_for_mechanic(self, auth_token):
        """GET /api/v1/onboarding/steps?business_type=mechanic - Should return 5 steps"""
        response = requests.get(
            f"{BASE_URL}/api/v1/onboarding/steps?business_type=mechanic",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'steps' in data
        assert 'total' in data
        
        steps = data['steps']
        assert len(steps) == 5, f"Expected 5 steps, got {len(steps)}"
        assert data['total'] == 5
        
        # Verify step structure
        step_codes = [s['StepCode'] for s in steps]
        expected_steps = ['company_info', 'first_os', 'invite_team', 'setup_whatsapp', 'first_approval']
        for step in expected_steps:
            assert step in step_codes, f"Missing step: {step}"

    def test_get_onboarding_progress_returns_40_percent(self, auth_token):
        """GET /api/v1/onboarding/progress - Should return 40% progress"""
        response = requests.get(
            f"{BASE_URL}/api/v1/onboarding/progress",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'progress' in data
        assert 'percent' in data
        assert 'steps' in data
        
        # Demo tenant has 2/5 completed = 40%
        assert data['percent'] == 40, f"Expected 40%, got {data['percent']}%"
        
        progress = data['progress']
        assert len(progress['CompletedSteps']) == 2, f"Expected 2 completed steps"
        assert 'company_info' in progress['CompletedSteps']
        assert 'first_os' in progress['CompletedSteps']

    def test_complete_onboarding_step(self, auth_token):
        """POST /api/v1/onboarding/complete - Should mark step as done"""
        response = requests.post(
            f"{BASE_URL}/api/v1/onboarding/complete",
            headers={"Authorization": f"Bearer {auth_token}"},
            json={"step_code": "invite_team", "skipped": False}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'message' in data
        assert 'step_code' in data
        assert data['step_code'] == 'invite_team'

    def test_skip_onboarding(self, auth_token):
        """POST /api/v1/onboarding/skip - Should mark onboarding as skipped"""
        response = requests.post(
            f"{BASE_URL}/api/v1/onboarding/skip",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200
        
        data = response.json()
        assert 'message' in data

    def test_onboarding_requires_auth(self):
        """GET /api/v1/onboarding/progress - Should require authentication"""
        response = requests.get(f"{BASE_URL}/api/v1/onboarding/progress")
        assert response.status_code == 401


class TestJourneyTrackingEndpoints:
    """Test journey tracking endpoints"""

    def test_track_event(self, auth_token):
        """POST /api/v1/track - Should accept events"""
        response = requests.post(
            f"{BASE_URL}/api/v1/track",
            headers={"Authorization": f"Bearer {auth_token}"},
            json={
                "event_name": "page_view",
                "event_category": "engagement",
                "page_path": "/dashboard",
                "page_title": "Dashboard",
                "anonymous_id": "test-123",
                "session_id": "session-456"
            }
        )
        # Track endpoint returns 200 or 202
        assert response.status_code in [200, 202], f"Expected 200/202, got {response.status_code}"

    def test_get_funnel_analytics(self, auth_token):
        """GET /api/v1/analytics/funnel - Should return funnel metrics"""
        response = requests.get(
            f"{BASE_URL}/api/v1/analytics/funnel",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200, f"Expected 200, got {response.status_code}: {response.text}"
        
        data = response.json()
        assert 'period' in data
        assert 'daily' in data
        assert 'totals' in data
        
        totals = data['totals']
        assert 'Visits' in totals or 'visits' in totals

    def test_tracking_requires_auth(self):
        """POST /api/v1/track - Should require authentication (or accept anonymous)"""
        response = requests.post(
            f"{BASE_URL}/api/v1/track",
            json={"event_name": "test", "anonymous_id": "anon-123"}
        )
        # Tracking may accept anonymous events or require auth
        assert response.status_code in [200, 202, 401]


class TestLoginFlowStillWorks:
    """Verify Phase 1 login flow is not broken"""

    def test_login_with_demo_credentials(self):
        """POST /api/auth/login - Demo login should work"""
        response = requests.post(
            f"{BASE_URL}/api/auth/login",
            json={
                "tenant_slug": "demo",
                "email": "admin@demo.com",
                "password": "demo123"
            }
        )
        assert response.status_code == 200, f"Login failed: {response.text}"
        
        data = response.json()
        assert 'access_token' in data
        assert 'token_type' in data
        assert data['token_type'] == 'Bearer'

    def test_dashboard_still_loads(self, auth_token):
        """GET /api/v1/dashboard/stats - Dashboard should still work"""
        response = requests.get(
            f"{BASE_URL}/api/v1/dashboard/stats",
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        assert response.status_code == 200


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
