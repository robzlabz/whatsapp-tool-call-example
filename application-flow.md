# WhatsApp AI Bot Application Flow

```mermaid
graph TD
    %% External Systems
    WA[WhatsApp User] --> |sends message| FONNTE[Fonnte API]
    FONNTE --> |webhook| APP[WhatsApp AI Bot]
    
    %% Application Entry Point
    APP --> |POST /webhook/fonnte| WEBHOOK[Webhook Handler]
    
    %% Main Application Components
    subgraph "Application Core"
        WEBHOOK --> |parse webhook| PROCESS[Process Message]
        PROCESS --> |async goroutine| MSG_HANDLER[Message Handler]
        
        MSG_HANDLER --> DB[(Database)]
        MSG_HANDLER --> OPENAI[OpenAI Service]
        MSG_HANDLER --> TOOLS[Tool Manager]
        MSG_HANDLER --> FONNTE_SVC[Fonnte Service]
    end
    
    %% Database Operations
    subgraph "Database Layer"
        DB --> |get/create| CONV[Conversation]
        DB --> |retrieve| HISTORY[Message History]
        DB --> |save| NEW_MSG[New Messages]
        DB --> |log| TOOL_EXEC[Tool Executions]
    end
    
    %% Message Processing Flow
    subgraph "Message Processing"
        MSG_HANDLER --> |1. Get conversation| CONV
        MSG_HANDLER --> |2. Build context| CONTEXT[Build Chat Context]
        CONTEXT --> |recent messages| HISTORY
        CONTEXT --> |system prompt| SYS_PROMPT[System Message]
        CONTEXT --> |current message| CURR_MSG[Current User Message]
    end
    
    %% OpenAI Integration
    subgraph "AI Processing"
        CONTEXT --> |chat messages + tools| OPENAI
        OPENAI --> |API call| OPENAI_API[OpenAI API]
        OPENAI_API --> |response| AI_RESPONSE[AI Response]
        AI_RESPONSE --> |check for| TOOL_CALLS{Has Tool Calls?}
    end
    
    %% Tool Execution Flow
    subgraph "Tool System"
        TOOL_CALLS --> |yes| TOOLS
        TOOLS --> |execute| IMG_TOOL[Image Generation Tool]
        IMG_TOOL --> |DALL-E API| DALLE[DALL-E API]
        DALLE --> |image URL| IMG_RESULT[Image Result]
        IMG_RESULT --> |send image| FONNTE_SVC
    end
    
    %% Response Handling
    subgraph "Response Flow"
        TOOL_CALLS --> |no| TEXT_RESP[Text Response]
        TEXT_RESP --> |send message| FONNTE_SVC
        FONNTE_SVC --> |API call| FONNTE_API[Fonnte API]
        FONNTE_API --> |deliver| WA
    end
    
    %% Data Persistence
    subgraph "Data Storage"
        MSG_HANDLER --> |save user msg| NEW_MSG
        MSG_HANDLER --> |save AI response| NEW_MSG
        MSG_HANDLER --> |update conversation| CONV
        TOOLS --> |log execution| TOOL_EXEC
    end
    
    %% Health & Monitoring
    subgraph "Monitoring"
        APP --> |GET /health| HEALTH[Health Check]
        APP --> |GET /stats| STATS[Statistics]
        HEALTH --> |200 OK| MONITOR[Monitoring System]
        STATS --> |metrics| MONITOR
    end
    
    %% Configuration & Services
    subgraph "Application Startup"
        MAIN[main.go] --> |load| CONFIG[Configuration]
        MAIN --> |initialize| LOGGER[Logger]
        MAIN --> |connect| DB
        MAIN --> |setup| SERVICES[Services]
        SERVICES --> FONNTE_SVC
        SERVICES --> OPENAI
        SERVICES --> TOOLS
        MAIN --> |start| HTTP_SERVER[HTTP Server]
        HTTP_SERVER --> APP
    end
    
    %% Error Handling
    subgraph "Error Handling"
        MSG_HANDLER --> |on error| ERROR_MSG[Error Message]
        ERROR_MSG --> |send error| FONNTE_SVC
        TOOLS --> |on failure| TOOL_ERROR[Tool Error]
        TOOL_ERROR --> |log & notify| ERROR_MSG
    end
    
    %% Styling
    classDef external fill:#e1f5fe
    classDef core fill:#f3e5f5
    classDef database fill:#e8f5e8
    classDef ai fill:#fff3e0
    classDef tools fill:#fce4ec
    classDef response fill:#e0f2f1
    
    class WA,FONNTE,OPENAI_API,DALLE,FONNTE_API external
    class APP,WEBHOOK,PROCESS,MSG_HANDLER core
    class DB,CONV,HISTORY,NEW_MSG,TOOL_EXEC database
    class OPENAI,AI_RESPONSE,TOOL_CALLS,CONTEXT ai
    class TOOLS,IMG_TOOL,IMG_RESULT tools
    class TEXT_RESP,FONNTE_SVC,ERROR_MSG response
```

## Application Flow Description

### 1. **Message Reception**
- WhatsApp user sends message through Fonnte API
- Fonnte delivers message via webhook to `/webhook/fonnte`
- Webhook handler parses and validates incoming message

### 2. **Message Processing**
- Message processing runs asynchronously in goroutine
- System retrieves or creates conversation record
- Builds chat context with recent message history
- Adds system prompt and current user message

### 3. **AI Processing**
- Sends chat context to OpenAI with available tools
- OpenAI processes message and determines response
- Checks if response includes tool calls (e.g., image generation)

### 4. **Tool Execution** (if applicable)
- Tool Manager executes requested tools
- Image generation tool calls DALL-E API
- Tool execution results are logged to database
- Generated content is prepared for delivery

### 5. **Response Delivery**
- Text responses sent directly via Fonnte
- Image responses include generated image URL and caption
- All messages delivered back to WhatsApp user

### 6. **Data Persistence**
- User messages saved to database
- AI responses saved to database
- Conversation metadata updated
- Tool execution logs maintained

### 7. **Error Handling**
- Comprehensive error handling at each step
- Error messages sent to user when issues occur
- All errors logged for debugging and monitoring

### 8. **Monitoring & Health**
- Health check endpoint for service monitoring
- Statistics endpoint for performance metrics
- Structured logging throughout application