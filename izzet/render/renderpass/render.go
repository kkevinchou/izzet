package renderpass

// func renderModels(shader *shaders.ShaderProgram,
// 	viewerContext context.ViewerContext,
// 	lightContext context.LightContext,
// 	renderContext context.RenderContext,
// 	renderableEntities []*entities.Entity,
// 	app renderiface.App,
// ) {
// 	shader.Use()

// 	if app.RuntimeConfig().FogEnabled {
// 		shader.SetUniformInt("fog", 1)
// 	} else {
// 		shader.SetUniformInt("fog", 0)
// 	}

// 	var fog int32 = 0
// 	if r.app.RuntimeConfig().FogDensity != 0 {
// 		fog = 1
// 	}
// 	shader.SetUniformInt("fog", fog)
// 	shader.SetUniformInt("fogDensity", app.RuntimeConfig().FogDensity)

// 	shader.SetUniformInt("width", int32(renderContext.Width()))
// 	shader.SetUniformInt("height", int32(renderContext.Height()))
// 	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
// 	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
// 	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
// 	shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
// 	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
// 	shader.SetUniformFloat("ambientFactor", app.RuntimeConfig().AmbientFactor)
// 	shader.SetUniformInt("shadowMap", 31)
// 	shader.SetUniformInt("depthCubeMap", 30)
// 	shader.SetUniformInt("cameraDepthMap", 29)
// 	shader.SetUniformInt("ambientOcclusion", 28)
// 	if app.RuntimeConfig().EnableSSAO {
// 		shader.SetUniformInt("enableAmbientOcclusion", 1)
// 	} else {
// 		shader.SetUniformInt("enableAmbientOcclusion", 0)
// 	}

// 	shader.SetUniformFloat("near", app.RuntimeConfig().Near)
// 	shader.SetUniformFloat("far", app.RuntimeConfig().Far)
// 	shader.SetUniformFloat("bias", app.RuntimeConfig().PointLightBias)
// 	if len(lightContext.PointLights) > 0 {
// 		shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
// 	}
// 	shader.SetUniformInt("hasColorOverride", 0)

// 	setupLightingUniforms(shader, lightContext.Lights)

// 	gl.ActiveTexture(gl.TEXTURE28)
// 	gl.BindTexture(gl.TEXTURE_2D, r.ssaoBlurTexture)

// 	gl.ActiveTexture(gl.TEXTURE29)
// 	gl.BindTexture(gl.TEXTURE_2D, r.cameraDepthTexture)

// 	gl.ActiveTexture(gl.TEXTURE30)
// 	gl.BindTexture(gl.TEXTURE_CUBE_MAP, r.depthCubeMapTexture)

// 	gl.ActiveTexture(gl.TEXTURE31)
// 	gl.BindTexture(gl.TEXTURE_2D, r.shadowMap.DepthTexture())

// 	var entityCount int
// 	for _, entity := range renderableEntities {
// 		if entity == nil || entity.MeshComponent == nil || !entity.MeshComponent.Visible {
// 			continue
// 		}

// 		// if app.RuntimeConfig().BatchRenderingEnabled && len(r.batchRenders) > 0 && entity.Static {
// 		// 	continue
// 		// }

// 		if entity.MeshComponent.InvisibleToPlayerOwner && app.GetPlayerEntity().GetID() == entity.GetID() {
// 			continue
// 		}

// 		entityCount++

// 		shader.SetUniformUInt("entityID", uint32(entity.ID))

// 		r.drawModel(
// 			shader,
// 			entity,
// 		)
// 	}

// 	app.MetricsRegistry().Inc("draw_entity_count", float64(entityCount))

// 	// if app.RuntimeConfig().BatchRenderingEnabled && len(r.batchRenders) > 0 {
// 	// 	shader.SetUniformInt("hasColorOverride", 0)
// 	// 	shader = r.shaderManager.GetShaderProgram("batch")
// 	// 	shader.Use()

// 	// 	if app.RuntimeConfig().FogEnabled {
// 	// 		shader.SetUniformInt("fog", 1)
// 	// 	} else {
// 	// 		shader.SetUniformInt("fog", 0)
// 	// 	}

// 	// 	var fog int32 = 0
// 	// 	if app.RuntimeConfig().FogDensity != 0 {
// 	// 		fog = 1
// 	// 	}
// 	// 	shader.SetUniformInt("fog", fog)
// 	// 	shader.SetUniformInt("fogDensity", app.RuntimeConfig().FogDensity)

// 	// 	shader.SetUniformInt("width", int32(renderContext.Width()))
// 	// 	shader.SetUniformInt("height", int32(renderContext.Height()))
// 	// 	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
// 	// 	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
// 	// 	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
// 	// 	shader.SetUniformFloat("shadowDistance", float32(r.shadowMap.ShadowDistance()))
// 	// 	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
// 	// 	shader.SetUniformFloat("ambientFactor", app.RuntimeConfig().AmbientFactor)
// 	// 	shader.SetUniformInt("shadowMap", 31)
// 	// 	shader.SetUniformInt("depthCubeMap", 30)
// 	// 	shader.SetUniformInt("cameraDepthMap", 29)
// 	// 	shader.SetUniformInt("ambientOcclusion", 28)
// 	// 	if app.RuntimeConfig().EnableSSAO {
// 	// 		shader.SetUniformInt("enableAmbientOcclusion", 1)
// 	// 	} else {
// 	// 		shader.SetUniformInt("enableAmbientOcclusion", 0)
// 	// 	}

// 	// 	shader.SetUniformFloat("near", app.RuntimeConfig().Near)
// 	// 	shader.SetUniformFloat("far", app.RuntimeConfig().Far)
// 	// 	shader.SetUniformFloat("bias", app.RuntimeConfig().PointLightBias)
// 	// 	if len(lightContext.PointLights) > 0 {
// 	// 		shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
// 	// 	}
// 	// 	shader.SetUniformInt("hasColorOverride", 0)

// 	// 	setupLightingUniforms(shader, lightContext.Lights)

// 	// 	gl.ActiveTexture(gl.TEXTURE28)
// 	// 	gl.BindTexture(gl.TEXTURE_2D, r.ssaoBlurTexture)

// 	// 	gl.ActiveTexture(gl.TEXTURE29)
// 	// 	gl.BindTexture(gl.TEXTURE_2D, r.cameraDepthTexture)

// 	// 	gl.ActiveTexture(gl.TEXTURE30)
// 	// 	gl.BindTexture(gl.TEXTURE_CUBE_MAP, r.depthCubeMapTexture)

// 	// 	gl.ActiveTexture(gl.TEXTURE31)
// 	// 	gl.BindTexture(gl.TEXTURE_2D, r.shadowMap.DepthTexture())

// 	// 	r.drawBatches(shader)
// 	// 	app.MetricsRegistry().Inc("draw_entity_count", 1)
// 	// }
// }

// // i considered using uniform blocks but the memory layout management seems like a huge pain
// // https://stackoverflow.com/questions/38172696/should-i-ever-use-a-vec3-inside-of-a-uniform-buffer-or-shader-storage-buffer-o
// func setupLightingUniforms(shader *shaders.ShaderProgram, lights []*entities.Entity) {
// 	if len(lights) > settings.MaxLightCount {
// 		panic(fmt.Sprintf("light count of %d exceeds max %d", len(lights), settings.MaxLightCount))
// 	}

// 	shader.SetUniformInt("lightCount", int32(len(lights)))
// 	for i, light := range lights {
// 		lightInfo := light.LightInfo

// 		diffuse := lightInfo.IntensifiedDiffuse()

// 		shader.SetUniformInt(fmt.Sprintf("lights[%d].type", i), int32(lightInfo.Type))
// 		shader.SetUniformVec3(fmt.Sprintf("lights[%d].dir", i), lightInfo.Direction3F)
// 		shader.SetUniformVec3(fmt.Sprintf("lights[%d].diffuse", i), diffuse)
// 		shader.SetUniformVec3(fmt.Sprintf("lights[%d].position", i), utils.Vec3F64ToF32(light.Position()))
// 		shader.SetUniformFloat(fmt.Sprintf("lights[%d].range", i), lightInfo.Range)
// 	}
// }
