import FormRecipient from "./generic/FormRecipient.js";

export default {
    name: 'SendMessage',
    components: {
        FormRecipient
    },
    data() {
        return {
            type: window.TYPEUSER,
            phone: '',
            text: '',
            reply_message_id: '',
            is_forwarded: false,
            mention_everyone: false,
            duration: 0,
            loading: false,
            selected_devices: [],
        }
    },
    computed: {
        phone_id() {
            return this.phone + this.type;
        },
    },
    methods: {
        openModal() {
            $('#modalSendMessage').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        isShowReplyId() {
            return this.type !== window.TYPESTATUS;
        },
        isShowBroadcastCheckbox() {
            return this.type === window.TYPESTATUS;
        },
        isGroup() {
            return this.type === window.TYPEGROUP;
        },
        getConnectedDevices() {
            return this.$root.connected_devices || this.$root.deviceList || [];
        },
        isAllDevicesSelected() {
            const devices = this.getConnectedDevices();
            if (devices.length === 0) return false;
            return devices.every(d => this.selected_devices.includes(d.id || d.device));
        },
        toggleAllDevices() {
            const devices = this.getConnectedDevices();
            if (this.isAllDevicesSelected()) {
                this.selected_devices = [];
            } else {
                this.selected_devices = devices.map(d => d.id || d.device);
            }
        },
        isValidForm() {
            // Validate phone number is not empty except for status type
            const isPhoneValid = this.type === window.TYPESTATUS || this.phone.trim().length > 0;

            // Validate message is not empty and has reasonable length
            const isMessageValid = this.text.trim().length > 0 && this.text.length <= 4096;

            return isPhoneValid && isMessageValid
        },
        async handleSubmit() {
            // Add validation check here to prevent submission when form is invalid
            if (!this.isValidForm() || this.loading) {
                return;
            }

            if (this.selected_devices.length > 0 && this.type === window.TYPESTATUS) {
                await this.handleBroadcastSubmit();
                return;
            }

            try {
                const response = await this.submitApi();
                showSuccessInfo(response);
                $('#modalSendMessage').modal('hide');
            } catch (err) {
                showErrorInfo(err);
            }
        },
        async handleBroadcastSubmit() {
            this.loading = true;
            try {
                // Use selected devices only
                const allDevices = this.$root.connected_devices || this.$root.deviceList || [];
                const targets = allDevices.filter(d => this.selected_devices.includes(d.id || d.device));
                console.log("Broadcast targets:", targets);

                if (targets.length === 0) {
                    throw new Error("No devices selected to broadcast to.");
                }
                const payload = {
                    phone: this.phone_id,
                    message: this.text.trim(),
                    is_forwarded: this.is_forwarded
                };
                if (this.duration && this.duration > 0) {
                    payload.duration = this.duration;
                }

                const promises = targets.map(device => {
                    const deviceId = device.id || device.device;
                    return window.http.post('/send/message', payload, {
                        headers: { 'X-Device-Id': deviceId }
                    }).then(() => ({ status: 'fulfilled', id: deviceId }))
                        .catch(err => ({ status: 'rejected', id: deviceId, error: err }));
                });

                const results = await Promise.all(promises);
                const success = results.filter(r => r.status === 'fulfilled');
                const failed = results.filter(r => r.status === 'rejected');

                if (success.length > 0) {
                    showSuccessInfo(`Success sent to ${success.length} devices.`);
                }
                if (failed.length > 0) {
                    showErrorInfo(`Failed to send to ${failed.length} devices.`);
                    console.error("Broadcast errors:", failed);
                }

                this.handleReset();
                $('#modalSendMessage').modal('hide');
            } catch (err) {
                showErrorInfo(err.message || err);
            } finally {
                this.loading = false;
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                const payload = {
                    phone: this.phone_id,
                    message: this.text.trim(),
                    is_forwarded: this.is_forwarded
                };
                if (this.reply_message_id !== '') {
                    payload.reply_message_id = this.reply_message_id;
                }

                if (this.duration && this.duration > 0) {
                    payload.duration = this.duration;
                }

                // Add mentions if mention_everyone is checked (only for groups)
                if (this.mention_everyone && this.type === window.TYPEGROUP) {
                    payload.mentions = ["@everyone"];
                }

                const response = await window.http.post('/send/message', payload);
                this.handleReset();
                return response.data.message;
            } catch (error) {
                if (error.response?.data?.message) {
                    throw new Error(error.response.data.message);
                }
                throw error;
            } finally {
                this.loading = false;
            }
        },
        handleReset() {
            this.phone = '';
            this.text = '';
            this.reply_message_id = '';
            this.is_forwarded = false;
            this.mention_everyone = false;
            this.duration = 0;
            this.selected_devices = [];
        },
    },
    template: `
    <div class="blue card" @click="openModal()" style="cursor: pointer">
        <div class="content">
            <a class="ui blue right ribbon label">Send</a>
            <div class="header">Send Message</div>
            <div class="description">
                Send any message to user or group
            </div>
        </div>
    </div>
    
    <!--  Modal SendMessage  -->
    <div class="ui small modal" id="modalSendMessage">
        <i class="close icon"></i>
        <div class="header">
            Send Message
        </div>
        <div class="content">
            <form class="ui form">
                <FormRecipient v-model:type="type" v-model:phone="phone" :show-status="true"/>
                <div class="field" v-if="isShowBroadcastCheckbox()">
                    <label>Select Devices to Send Status</label>
                    <div class="ui segment" style="max-height: 200px; overflow-y: auto;">
                        <div class="ui checkbox" style="display: block; margin-bottom: 10px; padding-bottom: 10px; border-bottom: 1px solid #eee;">
                            <input type="checkbox" :checked="isAllDevicesSelected()" @change="toggleAllDevices">
                            <label style="font-weight: bold;">Select All Devices</label>
                        </div>
                        <div v-for="device in getConnectedDevices()" :key="device.id || device.device" class="ui checkbox" style="display: block; margin-bottom: 8px;">
                            <input type="checkbox" :value="device.id || device.device" v-model="selected_devices">
                            <label>
                                <i class="mobile alternate icon"></i>
                                {{ device.name || device.pushname || device.id || device.device }}
                                <span class="ui mini label" :class="device.state === 'logged_in' ? 'green' : 'grey'">{{ device.state }}</span>
                            </label>
                        </div>
                        <div v-if="getConnectedDevices().length === 0" class="ui message">
                            <p>No devices connected. Please connect a device first.</p>
                        </div>
                    </div>
                    <div v-if="selected_devices.length > 0" class="ui mini label blue" style="margin-top: 8px;">
                        {{ selected_devices.length }} device(s) selected
                    </div>
                </div>
                <div class="field" v-if="isShowReplyId()">
                    <label>Reply Message ID</label>
                    <input v-model="reply_message_id" type="text"
                           placeholder="Optional: 57D29F74B7FC62F57D8AC2C840279B5B/3EB0288F008D32FCD0A424"
                           aria-label="reply_message_id">
                </div>
                <div class="field">
                    <label>Message</label>
                    <textarea v-model="text" placeholder="Hello this is message text"
                              aria-label="message"></textarea>
                </div>
                <div class="field" v-if="isShowReplyId()">
                    <label>Is Forwarded</label>
                    <div class="ui toggle checkbox">
                        <input type="checkbox" aria-label="is forwarded" v-model="is_forwarded">
                        <label>Mark message as forwarded</label>
                    </div>
                </div>
                <div class="field" v-if="isGroup()">
                    <label>Mention Everyone</label>
                    <div class="ui toggle checkbox">
                        <input type="checkbox" aria-label="mention everyone" v-model="mention_everyone">
                        <label>Mention all group participants (@everyone)</label>
                    </div>
                </div>
                <div class="field">
                    <label>Disappearing Duration (seconds)</label>
                    <input v-model.number="duration" type="number" min="0" placeholder="0 (no expiry)" aria-label="duration"/>
                </div>
            </form>
        </div>
        <div class="actions">
            <button class="ui approve positive right labeled icon button" 
                 :class="{'disabled': !isValidForm() || loading}"
                 @click.prevent="handleSubmit">
                Send
                <i class="send icon"></i>
            </button>
        </div>
    </div>
    `
}