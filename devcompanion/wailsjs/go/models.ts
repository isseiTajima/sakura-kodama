export namespace config {
	
	export class Config {
	    name: string;
	    tone: string;
	    encourage_freq: string;
	    monologue: boolean;
	    always_on_top: boolean;
	    mute: boolean;
	    model: string;
	    log_path: string;
	    scale: number;
	    click_through: boolean;
	    independent_window_opacity: number;
	    ollama_endpoint: string;
	    llm_backend: string;
	    anthropic_api_key: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.tone = source["tone"];
	        this.encourage_freq = source["encourage_freq"];
	        this.monologue = source["monologue"];
	        this.always_on_top = source["always_on_top"];
	        this.mute = source["mute"];
	        this.model = source["model"];
	        this.log_path = source["log_path"];
	        this.scale = source["scale"];
	        this.click_through = source["click_through"];
	        this.independent_window_opacity = source["independent_window_opacity"];
	        this.ollama_endpoint = source["ollama_endpoint"];
	        this.llm_backend = source["llm_backend"];
	        this.anthropic_api_key = source["anthropic_api_key"];
	    }
	}

}

export namespace monitor {
	
	export class MonitorEvent {
	    State: string;
	    Task: string;
	    Mood: string;
	
	    static createFrom(source: any = {}) {
	        return new MonitorEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.State = source["State"];
	        this.Task = source["Task"];
	        this.Mood = source["Mood"];
	    }
	}

}

