<script lang="ts">
    import { Modal, Heading, Button } from "flowbite-svelte";

    let { open = $bindable(false) } = $props();

    function closeModal() {
        open = false;
    }
</script>

<Modal bind:open outsideclose={true} class="w-full max-w-4xl">
    <div class="p-6">
        <Heading tag="h2" class="text-2xl font-bold mb-6">Privacy Policy</Heading>
        
        <div class="prose max-w-none text-sm text-gray-700 space-y-4 max-h-96 overflow-y-auto">
            <section>
                <h3 class="text-lg font-semibold mb-2">1. Information We Collect</h3>
                <p>
                    We collect information you provide directly to us, such as when you create an account, 
                    participate in contests, or contact us for support.
                </p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>Personal information (name, email address, profile information)</li>
                    <li>Contest submissions and related content</li>
                    <li>Communication preferences</li>
                    <li>Usage data and analytics</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">2. How We Use Your Information</h3>
                <p>We use the information we collect to:</p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>Provide, maintain, and improve our services</li>
                    <li>Process contest entries and manage competitions</li>
                    <li>Send you technical notices and support messages</li>
                    <li>Communicate with you about contests, features, and updates</li>
                    <li>Monitor and analyze trends and usage</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">3. Information Sharing</h3>
                <p>
                    We do not sell, trade, or otherwise transfer your personal information to third parties without your consent, 
                    except in the following circumstances:
                </p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>With your explicit consent</li>
                    <li>To comply with legal obligations</li>
                    <li>To protect our rights and safety</li>
                    <li>In connection with a business transfer</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">4. Data Security</h3>
                <p>
                    We implement appropriate security measures to protect your personal information against unauthorized access, 
                    alteration, disclosure, or destruction. However, no method of transmission over the internet is 100% secure.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">5. Cookies and Tracking</h3>
                <p>
                    We use cookies and similar tracking technologies to enhance your experience on our platform. 
                    You can control cookie settings through your browser preferences.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">6. Your Rights</h3>
                <p>You have the right to:</p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>Access your personal information</li>
                    <li>Correct inaccurate information</li>
                    <li>Delete your account and associated data</li>
                    <li>Opt-out of marketing communications</li>
                    <li>Data portability</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">7. Children's Privacy</h3>
                <p>
                    Our service is not intended for children under 13 years of age. We do not knowingly collect 
                    personal information from children under 13.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">8. International Data Transfers</h3>
                <p>
                    Your information may be transferred to and processed in countries other than your own. 
                    We ensure appropriate safeguards are in place for such transfers.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">9. Changes to This Policy</h3>
                <p>
                    We may update this privacy policy from time to time. We will notify you of any changes by 
                    posting the new policy on this page and updating the effective date.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">10. Contact Us</h3>
                <p>
                    If you have any questions about this privacy policy, please contact us at privacy@qplayground.com.
                </p>
            </section>
        </div>

        <div class="flex justify-end mt-6 pt-4 border-t">
            <Button color="primary" onclick={closeModal}>
                I Understand
            </Button>
        </div>
    </div>
</Modal>