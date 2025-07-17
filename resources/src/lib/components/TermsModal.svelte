<script lang="ts">
    import { Modal, Heading, Button } from "flowbite-svelte";

    let { open = $bindable(false) } = $props();

    function closeModal() {
        open = false;
    }
</script>

<Modal bind:open outsideclose={true} class="w-full max-w-4xl">
    <div class="p-6">
        <Heading tag="h2" class="text-2xl font-bold mb-6">Terms of Use</Heading>
        
        <div class="prose max-w-none text-sm text-gray-700 space-y-4 max-h-96 overflow-y-auto">
            <section>
                <h3 class="text-lg font-semibold mb-2">1. Acceptance of Terms</h3>
                <p>
                    By accessing and using Deltechverse, you accept and agree to be bound by the terms and provision of this agreement. 
                    If you do not agree to abide by the above, please do not use this service.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">2. Use License</h3>
                <p>
                    Permission is granted to temporarily download one copy of Deltechverse per device for personal, 
                    non-commercial transitory viewing only. This is the grant of a license, not a transfer of title, and under this license you may not:
                </p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>modify or copy the materials</li>
                    <li>use the materials for any commercial purpose or for any public display</li>
                    <li>attempt to reverse engineer any software contained on Deltechverse</li>
                    <li>remove any copyright or other proprietary notations from the materials</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">3. Contest Participation</h3>
                <p>
                    Users may participate in contests hosted on the platform. By participating, you agree to:
                </p>
                <ul class="list-disc list-inside ml-4 mt-2">
                    <li>Follow all contest rules and guidelines</li>
                    <li>Submit original content that you own or have rights to use</li>
                    <li>Accept the decisions of contest judges as final</li>
                    <li>Allow your submissions to be displayed publicly on the platform</li>
                </ul>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">4. User Accounts</h3>
                <p>
                    You are responsible for maintaining the confidentiality of your account and password. 
                    You agree to accept responsibility for all activities that occur under your account.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">5. Privacy Policy</h3>
                <p>
                    Your privacy is important to us. Please review our Privacy Policy, which also governs your use of the service, 
                    to understand our practices.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">6. Disclaimer</h3>
                <p>
                    The materials on Deltechverse are provided on an 'as is' basis. Deltechverse makes no warranties, 
                    expressed or implied, and hereby disclaims and negates all other warranties including without limitation, 
                    implied warranties or conditions of merchantability, fitness for a particular purpose, 
                    or non-infringement of intellectual property or other violation of rights.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">7. Limitations</h3>
                <p>
                    In no event shall Deltechverse or its suppliers be liable for any damages (including, without limitation, 
                    damages for loss of data or profit, or due to business interruption) arising out of the use or inability 
                    to use the materials on Deltechverse, even if Deltechverse or a Deltechverse authorized representative has been 
                    notified orally or in writing of the possibility of such damage.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">8. Revisions and Errata</h3>
                <p>
                    The materials appearing on Deltechverse could include technical, typographical, or photographic errors. 
                    Deltechverse does not warrant that any of the materials on its website are accurate, complete, or current.
                </p>
            </section>

            <section>
                <h3 class="text-lg font-semibold mb-2">9. Governing Law</h3>
                <p>
                    These terms and conditions are governed by and construed in accordance with the laws and you irrevocably 
                    submit to the exclusive jurisdiction of the courts in that state or location.
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