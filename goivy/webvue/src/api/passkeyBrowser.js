export function createPasskeyBrowser(credentialNavigator = globalThis.navigator) {
  return {
    isSupported() {
      return Boolean(globalThis.PublicKeyCredential && credentialNavigator?.credentials);
    },

    async create(optionsResponse) {
      if (!this.isSupported()) {
        throw new Error('passkeys are not supported by this browser');
      }
      const publicKey = decodePublicKeyCreationOptions(optionsResponse.publicKey);
      const credential = await credentialNavigator.credentials.create({ publicKey });
      return serializeRegistrationCredential(credential);
    },

    async authenticate(optionsResponse) {
      if (!this.isSupported()) {
        throw new Error('passkeys are not supported by this browser');
      }
      const publicKey = decodePublicKeyRequestOptions(optionsResponse.publicKey);
      const credential = await credentialNavigator.credentials.get({ publicKey });
      return serializeAssertionCredential(credential);
    },
  };
}

export function decodePublicKeyCreationOptions(options) {
  return {
    ...options,
    challenge: base64URLToBuffer(options.challenge),
    user: {
      ...options.user,
      id: base64URLToBuffer(options.user.id),
    },
    excludeCredentials: (options.excludeCredentials || []).map(decodeCredentialDescriptor),
  };
}

export function decodePublicKeyRequestOptions(options) {
  return {
    ...options,
    challenge: base64URLToBuffer(options.challenge),
    allowCredentials: (options.allowCredentials || []).map(decodeCredentialDescriptor),
  };
}

function decodeCredentialDescriptor(descriptor) {
  return {
    ...descriptor,
    id: base64URLToBuffer(descriptor.id),
  };
}

function serializeRegistrationCredential(credential) {
  if (!credential) {
    throw new Error('passkey registration was canceled');
  }
  return {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64URL(credential.response.clientDataJSON),
      attestationObject: bufferToBase64URL(credential.response.attestationObject),
      transports: typeof credential.response.getTransports === 'function' ? credential.response.getTransports() : [],
    },
  };
}

function serializeAssertionCredential(credential) {
  if (!credential) {
    throw new Error('passkey authentication was canceled');
  }
  return {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64URL(credential.response.clientDataJSON),
      authenticatorData: bufferToBase64URL(credential.response.authenticatorData),
      signature: bufferToBase64URL(credential.response.signature),
      userHandle: credential.response.userHandle ? bufferToBase64URL(credential.response.userHandle) : '',
    },
  };
}

export function base64URLToBuffer(encoded) {
  const padded = `${encoded}${'='.repeat((4 - (encoded.length % 4)) % 4)}`;
  const base64 = padded.replace(/-/g, '+').replace(/_/g, '/');
  const raw = globalThis.atob(base64);
  const bytes = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    bytes[i] = raw.charCodeAt(i);
  }
  return bytes.buffer;
}

export function bufferToBase64URL(buffer) {
  const bytes = new Uint8Array(buffer);
  let raw = '';
  for (const byte of bytes) {
    raw += String.fromCharCode(byte);
  }
  return globalThis.btoa(raw).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '');
}
