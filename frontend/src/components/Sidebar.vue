<template>
  <nav :class="{active}">
    <template v-if="isLogged">
      <router-link
        class="action"
        to="/files/"
        :aria-label="$t('menu.PRODUCT')"
        :title="$t('menu.PRODUCT')"
      >
        <i class="material-icons">business_center</i>
        <span>{{ $t('menu.PRODUCT') }}</span>
      </router-link>
      <router-link
        class="action"
        to="/files/"
        :aria-label="$t('menu.SALES')"
        :title="$t('menu.SALES')"
      >
        <i class="material-icons">redeem</i>
        <span>{{ $t('menu.SALES') }}</span>
      </router-link>
      <router-link
        class="action"
        to="/files/"
        :aria-label="$t('menu.ORDER')"
        :title="$t('menu.ORDER')"
      >
        <i class="material-icons">payment</i>
        <span>{{ $t('menu.ORDER') }}</span>
      </router-link>
      <router-link
        class="action"
        to="/files/"
        :aria-label="$t('menu.INFO')"
        :title="$t('menu.INFO')"
      >
        <i class="material-icons">info</i>
        <span>{{ $t('menu.INFO') }}</span>
      </router-link>

      <!--
      <div v-if="user">
        <button
          @click="$store.commit('showHover', 'newDir')"
          class="action"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">create_new_folder</i>
          <span>{{ $t('sidebar.newFolder') }}</span>
        </button>

        <button
          @click="$store.commit('showHover', 'newFile')"
          class="action"
          :aria-label="$t('sidebar.newFile')"
          :title="$t('sidebar.newFile')"
        >
          <i class="material-icons">note_add</i>
          <span>{{ $t('sidebar.newFile') }}</span>
        </button>
      </div>
      -->
      <div>
        <router-link
          class="action"
          to="/settings"
          :aria-label="$t('menu.SYSTEM')"
          :title="$t('menu.SYSTEM')"
        >
          <i class="material-icons">build</i>
          <span>{{ $t('menu.SYSTEM') }}</span>
        </router-link>

        <button
          @click="logout"
          class="action"
          id="logout"
          :aria-label="$t('sidebar.logout')"
          :title="$t('sidebar.logout')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t('sidebar.logout') }}</span>
        </button>
      </div>
    </template>
    <template v-else>
      <router-link
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t('sidebar.login') }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t('sidebar.signup') }}</span>
      </router-link>
    </template>

    <p class="credits">
      <span>
        <span v-if="disableExternal">Shop Management</span>
        <a
          v-else
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/agreyfox/eshop"
        >{{ $t('menu.SYS_NAME') }}</a>
        <span>V{{ version }}</span>
      </span>
      <br />
      <span>
        <a @click="help">{{ $t('sidebar.help') }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import * as auth from "@/utils/auth";
import {
  version,
  signup,
  disableExternal,
  noAuth,
  authMethod
} from "@/utils/constants";

export default {
  name: "sidebar",
  computed: {
    ...mapState(["user"]),
    ...mapGetters(["isLogged"]),
    active() {
      return this.$store.state.show === "sidebar";
    },
    signup: () => signup,
    version: () => version,
    disableExternal: () => disableExternal,
    noAuth: () => noAuth,
    authMethod: () => authMethod
  },
  methods: {
    help() {
      this.$store.commit("showHover", "help");
    },
    logout: auth.logout
  }
};
</script>
