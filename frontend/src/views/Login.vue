<template>
  <div id="login">
    <form @submit="submit">
      <img :src="logoURL" alt="EShop Manager" />
      <h1>{{ name }}</h1>
      <div v-if="error !== ''" class="wrong">{{ error }}</div>

      <input
        class="input input--block"
        type="text"
        v-model="username"
        :placeholder="$t('login.username')"
      />
      <input
        class="input input--block"
        type="password"
        v-model="password"
        :placeholder="$t('login.password')"
      />

      <input class="button button--block" type="submit" :value="$t('login.submit')" />
    </form>
  </div>
</template>

<script>
import * as auth from "@/utils/auth";
import { name, logoURL, signup } from "@/utils/constants";

export default {
  name: "login",
  computed: {
    signup: () => signup,
    name: () => name,
    logoURL: () => logoURL
  },
  data: function() {
    return {
      createMode: false,
      error: "",
      username: "",
      password: "",
      passwordConfirm: ""
    };
  },
  mounted() {
    //  if (!recaptcha) return;
    console.log("I am here");
    /*  window.grecaptcha.render("recaptcha", {
      sitekey: recaptchaKey
    }); */
  },
  methods: {
    async submit(event) {
      event.preventDefault();
      event.stopPropagation();

      let redirect = this.$route.query.redirect;
      if (redirect === "" || redirect === undefined || redirect === null) {
        redirect = "/settings/";
      }

      try {
        await auth.login(this.username, this.password, this.username);
        this.$router.push({ path: redirect });
      } catch (e) {
        if (e.message == 409) {
          this.error = this.$t("login.usernameTaken");
        } else {
          this.error = this.$t("login.wrongCredentials");
        }
      }
    }
  }
};
</script>
