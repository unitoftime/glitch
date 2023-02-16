if (!WebAssembly.instantiateStreaming) {
  // polyfill
  WebAssembly.instantiateStreaming = (source, importObject) => {
    const instantiate = (buffer) =>
      WebAssembly.instantiate(buffer, importObject);

    if (source instanceof Response) {
      return source.arrayBuffer().then(instantiate);
    }

    return source.then((res) => res.arrayBuffer()).then(instantiate);
  };
}

const go = new Go();

WebAssembly.instantiateStreaming(fetch("wasm.wasm"), go.importObject)
  .then(async ({ module, instance }) => {
    console.clear();

    try {
      await go.run(instance);
      // Note: Don't know why we need to reset the instance here. We don't seem use it again.
      instance = await WebAssembly.instantiate(module, go.importObject);
      // Note: Logging the result here seems wrong.
      console.log("Ran WASM:", result);
    } catch (error) {
      console.log("Failed to run WASM:", error);
    }
  })
  .catch((error) => {
    console.log("Could not create wasm instance", error);
  });
